package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"code.google.com/p/gcfg"
	_ "github.com/go-sql-driver/mysql"
	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/utils"
)

type config struct {
	Mysql struct {
		Ip       string
		Port     int
		Username string
		Password string
		Db       string
		Table    string
	}
	Index struct {
		Begin int
		Count int
	}
}

var (
	cfg  config
	keys []string = []string{"one_key", "sec_key", "username", "role_id"}
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

	db, e := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Mysql.Username, cfg.Mysql.Password, cfg.Mysql.Ip, cfg.Mysql.Port, cfg.Mysql.Db))
	utils.CheckError(e)
	defer db.Close()

	s := fmt.Sprintf("INSERT INTO %s(", cfg.Mysql.Table)
	sv := "("
	for i := 0; i < len(keys); i++ {
		if i > 0 {
			s += ","
			sv += ","
		}
		s += fmt.Sprintf("`%s`", keys[i])
		sv += "?"
	}
	s += ") VALUES"
	sv += "),"
	t := []string{}
	for row := 0; row < cfg.Index.Count; row++ {
		s += sv
		n := strconv.Itoa(cfg.Index.Begin + row)
		t = append(t, n)
		t = append(t, n)
		t = append(t, "")
		t = append(t, "0")
	}
	s = s[0 : len(s)-1]
	args := []interface{}{}
	for i := 0; i < len(t); i++ {
		args = append(args, t[i])
	}
	_, e = db.Exec(s, args...)
	utils.CheckError(e)

	if runtime.GOOS == "windows" {
		bufio.NewReader(os.Stdin).ReadLine()
	}
}
