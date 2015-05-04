package main

import (
	"database/sql"
	"fmt"
	"strings"

	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/utils"

	_ "github.com/go-sql-driver/mysql"
)

func dumpMysql(info *excel, base string) {
	if !checkForServer(info) {
		return
	}
	db, e := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Mysql.Username, cfg.Mysql.Password, cfg.Mysql.Ip, cfg.Mysql.Port, cfg.Mysql.Db))
	utils.CheckError(e)
	defer db.Close()

	fname := pathext.BaseWithoutExt(base)
	s := fmt.Sprintf("DROP TABLE IF EXISTS %s;", fname)
	if cfg.Other.LogV == 1 {
		fmt.Println(s)
	}
	_, e = db.Exec(s)
	utils.CheckError(e)

	s = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(", fname)
	for i := 0; i < len(info.keys); i++ {
		if !checkExtraForServer(info.extra[i]) {
			continue
		}
		if info.types[i] == "string" || strings.HasPrefix(info.types[i], "array_") {
			s += fmt.Sprintf("\n    `%s` text,", info.keys[i])
		} else {
			s += fmt.Sprintf("\n    `%s` %s,", info.keys[i], info.types[i])
		}
	}
	s += fmt.Sprintf("\n    PRIMARY KEY(`%s`)", info.keys[0])
	s += "\n) DEFAULT CHARACTER SET utf8;"
	if cfg.Other.LogV == 1 {
		fmt.Println(s)
	}
	_, e = db.Exec(s)
	utils.CheckError(e)

	s = fmt.Sprintf("INSERT INTO %s(", fname)
	sv := "("
	for i := 0; i < len(info.keys); i++ {
		if !checkExtraForServer(info.extra[i]) {
			continue
		}
		if i > 0 {
			s += ","
			sv += ","
		}
		s += fmt.Sprintf("`%s`", info.keys[i])
		sv += "?"
	}
	s += ") VALUES"
	sv += "),"
	if cfg.Other.LogV == 1 {
		fmt.Printf("%s%s\n", s, sv)
	}
	t := []string{}
	for _, dataRow := range info.data {
		s += sv
		for i := 0; i < len(dataRow); i++ {
			if !checkExtraForServer(info.extra[i]) {
				continue
			}
			dataCol := dataRow[i]
			t = append(t, dataCol["value"])
			if cfg.Other.LogV == 1 {
				fmt.Printf("%s ", dataCol["value"])
			}
		}
		if cfg.Other.LogV == 1 {
			fmt.Print("\n")
		}
	}
	s = s[0 : len(s)-1]
	args := []interface{}{}
	for i := 0; i < len(t); i++ {
		args = append(args, t[i])
	}
	_, e = db.Exec(s, args...)
	utils.CheckError(e)
}

func cleanupMysql() {
	db, e := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", cfg.Mysql.Username, cfg.Mysql.Password, cfg.Mysql.Ip, cfg.Mysql.Port))
	utils.CheckError(e)
	s := fmt.Sprintf("DROP DATABASE IF EXISTS %s", cfg.Mysql.Db)
	if cfg.Other.LogV == 1 {
		fmt.Println(s)
	}
	_, e = db.Exec(s)
	utils.CheckError(e)
	db.Close()
}
