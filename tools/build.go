// cd ../../../..
// export GOPATH=`pwd`
// go run ./src/uuzu.com/hanxl/tools/build.go
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"uuzu.com/hanxl/pathext"
)

var (
	packages = []string{
		"uuzu.com/hanxl/buildAndroid",
		"uuzu.com/hanxl/compress",
		"uuzu.com/hanxl/fileServe",
		"uuzu.com/hanxl/genProto",
		"uuzu.com/hanxl/genProtoGo",
		"uuzu.com/hanxl/genTestAccount",
		"uuzu.com/hanxl/publishSvn",
		"uuzu.com/hanxl/runtime",
		"uuzu.com/hanxl/uploadConfig",
	}
)

func main() {
	var pathBinOrigin, pathBin string
	gopath := os.Getenv("GOPATH")
	iniPathOrigin := filepath.Join(gopath, "src", "uuzu.com", "hanxl", "tools", "config")
	iniPath := filepath.Join(gopath, "bin")
	if runtime.GOOS == "windows" {
		pathBinOrigin = filepath.Join(gopath, "bin")
		pathBin = filepath.Join(pathBinOrigin, "windows_arm64")
		os.RemoveAll(pathBin)
		os.MkdirAll(pathBin, os.ModePerm)
	} else if runtime.GOOS == "linux" {
		pathBinOrigin = filepath.Join(gopath, "bin")
		pathBin = filepath.Join(gopath, "bin", "linux_arm64")
		os.RemoveAll(pathBin)
		os.MkdirAll(pathBin, os.ModePerm)
	}
	env := os.Environ()
	for _, v := range packages {
		// fmt.Println("windows", v)
		fmt.Println(v)
		cmd := exec.Command("go", "install", v)
		cmd.Env = env
		// cmd.Env = append(cmd.Env, "GOOS=windows")
		e := cmd.Run()
		checkError(e)

		arr := strings.Split(v, "/")
		binName := arr[len(arr)-1]

		iniName := binName + ".ini"
		iniOrigin := filepath.Join(iniPathOrigin, iniName)
		ini := filepath.Join(iniPath, iniName)
		if pathext.Exists(iniOrigin) && !pathext.Exists(ini) {
			r, _ := os.Open(iniOrigin)
			w, _ := os.Create(ini)
			defer r.Close()
			defer w.Close()
			io.Copy(w, r)
		}

		if runtime.GOOS == "windows" {
			binName += ".exe"
			os.Rename(filepath.Join(pathBinOrigin, binName), filepath.Join(pathBin, binName))
		} else if runtime.GOOS == "linux" {
			os.Rename(filepath.Join(pathBinOrigin, binName), filepath.Join(pathBin, binName))
		}

		// fmt.Println("darwin", v)
		// cmd = exec.Command("go", "install", v)
		// cmd.Env = env
		// cmd.Env = append(cmd.Env, "GOOS=darwin")
		// e = cmd.Run()
		// checkError(e)

		// fmt.Println("linux", v)
		// cmd = exec.Command("go", "install", v)
		// cmd.Env = env
		// cmd.Env = append(cmd.Env, "GOOS=linux")
		// e = cmd.Run()
		// checkError(e)
	}
}

func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
