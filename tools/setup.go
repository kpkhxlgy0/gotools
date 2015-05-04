// cd ../../../..
// export GOPATH=`pwd`
// go run ./src/uuzu.com/hanxl/tools/setup.go
package main

import (
	"fmt"
	"os"
	"os/exec"
)

var (
	packages = []string{
		"github.com/xxtea/xxtea-go/xxtea",
		"github.com/tealeg/xlsx",
		"github.com/go-sql-driver/mysql",
		"github.com/mattn/go-sqlite3",
		"code.google.com/p/gcfg",
	}
)

func main() {
	env := os.Environ()
	for _, v := range packages {
		// fmt.Println("windows", v)
		fmt.Println(v)
		cmd := exec.Command("go", "get", "-u", v)
		cmd.Env = env
		// cmd.Env = append(cmd.Env, "GOOS=windows")
		e := cmd.Run()
		checkError(e)

		// fmt.Println("darwin", v)
		// cmd = exec.Command("go", "get", "-u", v)
		// cmd.Env = env
		// cmd.Env = append(cmd.Env, "GOOS=darwin")
		// e = cmd.Run()
		// checkError(e)

		// fmt.Println("linux", v)
		// cmd = exec.Command("go", "get", "-u", v)
		// cmd.Env = env
		// cmd.Env = append(cmd.Env, "GOOS=linux")
		// e = cmd.Run()
		// checkError(e)
	}
}

func checkError(e error) {
	if e != nil {
		fmt.Println("WARNING:", "not support")
	}
}
