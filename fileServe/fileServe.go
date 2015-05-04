package main

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"uuzu.com/hanxl/pathext"
)

func main() {
	conn, err := net.Dial("udp", "www.baidu.com:80")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	arr := strings.Split(conn.LocalAddr().String(), ":")
	ip := arr[0]
	conn.Close()

	workdir := filepath.Join(pathext.GetExecDir(), "../../..")
	filedir := filepath.Join(workdir, "tools", "dist")
	h := http.FileServer(http.Dir(filedir))
	println(ip + ":8000")
	http.ListenAndServe(ip+":8000", h)
}
