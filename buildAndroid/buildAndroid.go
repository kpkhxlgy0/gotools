package main

import (
	"os/exec"
	"path/filepath"

	"uuzu.com/hanxl/pathext"
)

func main() {
	workdir := filepath.Join(pathext.GetExecDir(), "../../..")
	cmd := exec.Command("cocos", "compile", "-s", workdir, "-p", "android", "-j", "8")
	cmd.Run()
}
