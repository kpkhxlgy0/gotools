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

	"code.google.com/p/gcfg"
	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/utils"
)

type config struct {
	Url struct {
		In  string
		Out string
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

	cmd := exec.Command("svn", "rm", fmt.Sprintf("%s/nsxd", cfg.Url.Out), "-m", "\"x\"")
	e = cmd.Run()
	if e != nil {
		fmt.Println(cmd, "not done.")
	}

	cmd = exec.Command("svn", "mkdir", "--parents", fmt.Sprintf("%s/nsxd/tools", cfg.Url.Out), "-m", "\"x\"")
	e = cmd.Run()
	utils.CheckError(e)

	svnCopy("res")
	svnCopy("src")
	svnCopy("runtime")
	svnCopy("config.json")
	svnCopy("tools/gopath/bin/runtime")
	svnCopy("tools/gopath/bin/runtime.exe")

	if runtime.GOOS == "windows" {
		bufio.NewReader(os.Stdin).ReadLine()
	}
}

func svnCopy(dir string) {
	cmd := exec.Command("svn", "cp", fmt.Sprintf("%s/nsxd/%s", cfg.Url.In, dir), fmt.Sprintf("%s/nsxd/%s", cfg.Url.Out, dir), "-m", "\"x\"")
	e := cmd.Run()
	utils.CheckError(e)
}
