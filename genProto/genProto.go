package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"code.google.com/p/gcfg"
	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/protoParser"
	"uuzu.com/hanxl/utils"
)

type config struct {
	Dir struct {
		In  string
		Out string
	}
	Other struct {
		CleanupFirst int
	}
}

var (
	cfg config
)

func main() {
	execDir := pathext.GetExecDir()
	base := pathext.BaseWithoutExt(pathext.GetExecFile())
	iniFile := filepath.Join(execDir, base+".ini")
	fmt.Println(iniFile)
	if pathext.Exists(iniFile) == false {
		log.Fatal(errors.New("iniFile not found"))
	}
	e := gcfg.ReadFileInto(&cfg, iniFile)
	utils.CheckError(e)
	fmt.Printf("%+v\n", cfg)

	if pathext.Exists(cfg.Dir.Out) {
		if cfg.Other.CleanupFirst == 1 {
			os.RemoveAll(cfg.Dir.Out)
			os.MkdirAll(cfg.Dir.Out, os.ModePerm)
		}
	} else {
		os.MkdirAll(cfg.Dir.Out, os.ModePerm)
	}

	curDir, e := os.Getwd()
	os.Chdir(cfg.Dir.In)
	cmd := exec.Command("svn", "up")
	cmd.Run()
	os.Chdir(curDir)

	args := []string{}
	args = append(args, "-o"+filepath.Join(cfg.Dir.Out, "proto.pb"))
	args = append(args, "-I"+cfg.Dir.In)

	protos := &protoParser.Proto{}

	e = filepath.Walk(filepath.Join(cfg.Dir.In, "client"), func(path string, finfo os.FileInfo, err error) error {
		if finfo == nil {
			return err
		} else if finfo.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if strings.HasSuffix(base, ".proto") {
			args = append(args, path)
			proto := protoParser.Parse(path)
			if proto != nil {
				for _, v := range proto.Items {
					protos.Items = append(protos.Items, v)
				}
			}
		}
		return nil
	})
	utils.CheckError(e)

	cmd = exec.Command("protoc", args...)
	cmd.Run()

	dumpEnum(protos)
	opcode := protoParser.ParseOpcode(filepath.Join(cfg.Dir.In, "client", "opcode.proto"))
	dumpOpcode(opcode)

	if runtime.GOOS == "windows" {
		bufio.NewReader(os.Stdin).ReadLine()
	}
}
