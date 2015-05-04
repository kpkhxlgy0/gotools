package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xxtea/xxtea-go/xxtea"
	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/utils"
)

func dumpSqlite(info *excel, base string) {
	if !checkForClient(info) {
		return
	}
	fname := pathext.BaseWithoutExt(base)
	fpath := filepath.Join(cfg.Dir.Out, fname+".sqlite3")
	if pathext.Exists(fpath) {
		os.Remove(fpath)
	}
	db, e := sql.Open("sqlite3", fpath)
	utils.CheckError(e)
	defer db.Close()
	s := "CREATE TABLE main("
	if cfg.Crypto.Key == "" {
		s += fmt.Sprintf("`%s` INTEGER PRIMARY KEY", formatXXTEA(info.keys[0]))
	} else {
		s += fmt.Sprintf("`%s` TEXT PRIMARY KEY", formatXXTEA(info.keys[0]))
	}
	for i := 1; i < len(info.keys); i++ {
		if !checkExtraForClient(info.extra[i]) {
			continue
		}
		if cfg.Crypto.Key == "" {
			if info.types[i] == "int" {
				s += fmt.Sprintf(",`%s` INTEGER", formatXXTEA(info.keys[i]))
			} else if info.types[i] == "float" {
				s += fmt.Sprintf(",`%s` FLOAT", formatXXTEA(info.keys[i]))
			} else {
				s += fmt.Sprintf(",`%s` TEXT", formatXXTEA(info.keys[i]))
			}
		} else {
			s += fmt.Sprintf(",`%s` TEXT", formatXXTEA(info.keys[i]))
		}
	}
	s += ")"
	if cfg.Other.LogV == 1 {
		fmt.Println(s)
	}
	_, e = db.Exec(s)
	utils.CheckError(e)
	for _, dataRow := range info.data {
		s := "INSERT INTO main VALUES("
		sep := ""
		for i := 0; i < len(dataRow); i++ {
			if !checkExtraForClient(info.extra[i]) {
				continue
			}
			dataCol := dataRow[i]
			s += sep
			if sep == "" {
				sep = ","
			}
			t := dataCol["type"]
			d := dataCol["value"]
			if cfg.Crypto.Key == "" {
				if t == "int" || t == "float" {
					s += d
				} else {
					s += fmt.Sprintf("\"%s\"", d)
				}
			} else {
				s += fmt.Sprintf("\"%s\"", formatXXTEA(d))
			}
		}
		s += ")"
		if cfg.Other.LogV == 1 {
			fmt.Println(s)
		}
		_, e = db.Exec(s)
		utils.CheckError(e)
	}
}

func formatXXTEA(s string) string {
	if cfg.Crypto.Key == "" {
		return s
	}
	ss := xxtea.Encrypt([]byte(s), []byte(cfg.Crypto.Key))
	return base64.StdEncoding.EncodeToString(ss)
}
