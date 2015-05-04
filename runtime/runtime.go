package main

import (
	"os/exec"
	"path/filepath"
	"runtime"

	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/utils"
)

func main() {
	workdir := filepath.Join(pathext.GetExecDir(), "../../..")
	writablePath := filepath.Join(workdir, "writable-path")
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		exe := filepath.Join(workdir, "runtime/win32/nsxd.exe")
		cmd = exec.Command("cmd", "/C", "start "+exe, "-workdir", workdir, "-writable-path", writablePath)
	} else {
		exe := filepath.Join(workdir, "runtime/mac/nsxd Mac.app/Contents/MacOS/nsxd Mac")
		cmd = exec.Command(exe, "-workdir", workdir, "-writable-path", writablePath)
	}
	e := cmd.Start()
	utils.CheckError(e)
}
