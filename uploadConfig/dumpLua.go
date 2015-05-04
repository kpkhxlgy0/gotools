package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/utils"
)

func dumpLua(info *excel, base string) {
	if checkForClient(info) == false {
		return
	}
	fname := pathext.BaseWithoutExt(base)
	fpath := filepath.Join(cfg.Dir.Out, fname+".lua")
	if pathext.Exists(fpath) {
		os.Remove(fpath)
	}
	f, e := os.Create(fpath)
	defer f.Close()
	utils.CheckError(e)

	f.WriteString("return {\n")

	sepRow := ""
	for _, dataRow := range info.data {
		f.WriteString(sepRow)
		if sepRow == "" {
			sepRow = ",\n"
		}
		f.WriteString("{")
		sep := ""
		for i := 0; i < len(dataRow); i++ {
			if !checkExtraForClient(info.extra[i]) {
				continue
			}
			dataCol := dataRow[i]
			f.WriteString(sep)
			if sep == "" {
				sep = ","
			}
			k := dataCol["key"]
			t := dataCol["type"]
			d := dataCol["value"]
			f.WriteString(fmt.Sprintf("%s=", k))
			if t == "int" || t == "float" {
				if d == "" {
					f.WriteString("nil")
				} else {
					f.WriteString(d)
				}
			} else {
				d = strings.Replace(d, "\n", "\\n", -1)
				f.WriteString(d)
			}
		}
		f.WriteString("}")
	}

	f.WriteString("\n}")
}
