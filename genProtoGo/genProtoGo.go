package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
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

	e = filepath.Walk(filepath.Join(cfg.Dir.In, "client"), func(path string, finfo os.FileInfo, err error) error {
		if finfo == nil {
			return err
		} else if finfo.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if strings.HasSuffix(base, ".proto") {
			proto := protoParser.ParseGo(path)
			if proto != nil {
				dumpGolang(proto)
			}
		}
		return nil
	})
	utils.CheckError(e)

	if runtime.GOOS == "windows" {
		bufio.NewReader(os.Stdin).ReadLine()
	}
}
