// cd ../../../..
// export GOPATH=`pwd`
// go run ./src/uuzu.com/hanxl/tools/build.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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
	env := os.Environ()
	for _, v := range packages {
		fmt.Println("windows", v)
		cmd := exec.Command("go", "install", v)
		cmd.Env = env
		cmd.Env = append(cmd.Env, "GOOS=windows")
		e := cmd.Run()
		checkError(e)

		fmt.Println("darwin", v)
		cmd = exec.Command("go", "install", v)
		cmd.Env = env
		cmd.Env = append(cmd.Env, "GOOS=darwin")
		e = cmd.Run()
		checkError(e)

		fmt.Println("linux", v)
		cmd = exec.Command("go", "install", v)
		cmd.Env = env
		cmd.Env = append(cmd.Env, "GOOS=linux")
		e = cmd.Run()
		checkError(e)
	}
}

func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
