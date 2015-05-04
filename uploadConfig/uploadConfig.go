package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"code.google.com/p/gcfg"
	"github.com/tealeg/xlsx"
	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/utils"
)

type config struct {
	Dir struct {
		In  string
		Out string
	}
	Mysql struct {
		Ip       string
		Port     int
		Username string
		Password string
		Db       string
	}
	Crypto struct {
		Key string
	}
	Enable struct {
		Mysql  int
		Golang int
		Sqlite int
		Lua    int
	}
	Other struct {
		LogV         int
		CleanupFirst int
	}
}

type excelWithArray struct {
	desc    string
	key     string
	typeStr string
	extra   string
	index   []int
	index1  []int
}

type excel struct {
	desc  []string
	keys  []string
	types []string
	extra []string
	data  [][]map[string]string
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

	if pathext.Exists(cfg.Dir.Out) {
		if cfg.Other.CleanupFirst == 1 {
			os.RemoveAll(cfg.Dir.Out)
			os.MkdirAll(cfg.Dir.Out, os.ModePerm)
		}
	} else {
		os.MkdirAll(cfg.Dir.Out, os.ModePerm)
	}
	if cfg.Enable.Mysql == 1 && cfg.Other.CleanupFirst == 1 {
		cleanupMysql()
	}
	if cfg.Enable.Golang == 1 {
		cleanupMg()
	}

	e = filepath.Walk(cfg.Dir.In, func(path string, finfo os.FileInfo, err error) error {
		if finfo == nil {
			return err
		} else if finfo.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		pattern := regexp.MustCompile(`^\w+\.xlsx$`)
		if pattern.MatchString(base) {
			fmt.Println(path)
			info := readExcel(path)
			// fmt.Printf("%v\n", data)
			if cfg.Enable.Lua == 1 {
				dumpLua(info, base)
			}
			if cfg.Enable.Sqlite == 1 {
				dumpSqlite(info, base)
			}
			if cfg.Enable.Mysql == 1 {
				dumpMysql(info, base)
			}
			if cfg.Enable.Golang == 1 {
				dumpGolang(info, base)
			}
		}
		return nil
	})
	utils.CheckError(e)
	if cfg.Enable.Golang == 1 {
		dumpGolangMg()
	}

	if runtime.GOOS == "windows" {
		bufio.NewReader(os.Stdin).ReadLine()
	}
}

func readExcel(path string) *excel {
	xlFile, e := xlsx.OpenFile(path)
	utils.CheckError(e)
	sheet := xlFile.Sheets[0]
	if sheet.MaxRow < 5 || sheet.MaxCol < 1 {
		return nil
	}
	rows := sheet.Rows
	descRow := rows[0]
	keysRow := rows[1]
	typesRow := rows[2]
	extraRow := rows[3]

	temp := []*excelWithArray{}
	for i := 0; i < sheet.MaxCol; i++ {
		if keysRow.Cells[i].String() == "" {
			continue
		}
		arr := strings.Split(keysRow.Cells[i].String(), "_")
		key := arr[0]
		if len(arr) > 1 {
			var t *excelWithArray = nil
			for _, v := range temp {
				if v.key == key {
					t = v
					break
				}
			}
			if t == nil {
				t = new(excelWithArray)
				t.desc = descRow.Cells[i].String()
				t.key = key
				t.typeStr = "array_" + typesRow.Cells[i].String()
				t.extra = extraRow.Cells[i].String()
				temp = append(temp, t)
			}
			t.index = append(t.index, i)
			intValue, e := strconv.Atoi(arr[1])
			utils.CheckError(e)
			t.index1 = append(t.index1, intValue-1)
		} else {
			t := new(excelWithArray)
			t.desc = descRow.Cells[i].String()
			t.key = key
			t.typeStr = typesRow.Cells[i].String()
			t.extra = extraRow.Cells[i].String()
			t.index = append(t.index, i)
			temp = append(temp, t)
		}
	}

	ret := new(excel)
	for _, v := range temp {
		ret.desc = append(ret.desc, v.desc)
		ret.keys = append(ret.keys, v.key)
		ret.types = append(ret.types, v.typeStr)
		ret.extra = append(ret.extra, v.extra)
	}

	for i := 4; i < sheet.MaxRow; i++ {
		if sheet.Cell(i, 0).String() == "" {
			continue
		}
		row := []map[string]string{}
		for _, v := range temp {
			t := make(map[string]string)
			t["desc"] = v.desc
			t["key"] = v.key
			t["type"] = v.typeStr
			if len(v.index1) > 0 {
				if v.typeStr == "array_int" {
					value := []int{}
					for ii := 0; ii < len(v.index); ii++ {
						j := v.index[ii]
						index := v.index1[ii]
						for len(value) < index+1 {
							value = append(value, 0)
						}
						// if v.typeStr == "array_int" {
						intValue, e := sheet.Cell(i, j).Int()
						utils.CheckError(e)
						value[index] = intValue
						// }
					}
					buf, e := json.Marshal(value)
					utils.CheckError(e)
					t["value"] = string(buf)
				} else {
					value := []string{}
					for ii := 0; ii < len(v.index); ii++ {
						j := v.index[ii]
						index := v.index1[ii]
						for len(value) < index+1 {
							value = append(value, "")
						}
						value[index] = sheet.Cell(i, j).String()
					}
					buf, e := json.Marshal(value)
					utils.CheckError(e)
					t["value"] = string(buf)
				}
			} else {
				j := v.index[0]
				t["value"] = sheet.Cell(i, j).String()
			}
			row = append(row, t)
		}
		ret.data = append(ret.data, row)
	}
	return ret
}

func checkExtraForClient(v string) bool {
	return v != "skip" && v != "server"
}

func checkForClient(info *excel) bool {
	for _, v := range info.extra {
		if checkExtraForClient(v) {
			return true
		}
	}
	return false
}

func checkExtraForServer(v string) bool {
	return v != "skip" && v != "client"
}

func checkForServer(info *excel) bool {
	for _, v := range info.extra {
		if checkExtraForServer(v) {
			return true
		}
	}
	return false
}
