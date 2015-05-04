package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/utils"
)

var configList []string

func dumpGolang(info *excel, base string) {
	if !checkForServer(info) {
		return
	}
	fname := pathext.BaseWithoutExt(base)
	fpath := filepath.Join(cfg.Dir.Out, fname+"_config.go")
	if pathext.Exists(fpath) {
		os.Remove(fpath)
	}
	f, e := os.Create(fpath)
	defer f.Close()
	defer func() {
		configList = append(configList, utils.FormatName(fname)+"Config")
	}()
	utils.CheckError(e)

	f.WriteString(fmt.Sprintf(`//%s created by tool. Do not modify.
//date: %s
//author: hanxl
`, fname, time.Now().Format("2006-01-02 15:04:05")))
	f.WriteString(`
package config
import (
    "database/sql"
    "encoding/json"
    "fmt"
)
`)
	mName := fname
	mNameF := utils.FormatName(mName)
	f.WriteString(fmt.Sprintf("\ntype %s struct {\n", mNameF))
	for i := 0; i < len(info.keys); i++ {
		if !checkExtraForServer(info.extra[i]) {
			continue
		}
		f.WriteString(fmt.Sprintf("    // %s\n", info.desc[i]))
		switch info.types[i] {
		case "array_int":
			f.WriteString(fmt.Sprintf("    %s []int32\n", utils.FormatName(info.keys[i])))
		case "array_float":
			f.WriteString(fmt.Sprintf("    %s []float32\n", utils.FormatName(info.keys[i])))
		case "array_string":
			f.WriteString(fmt.Sprintf("    %s []string\n", utils.FormatName(info.keys[i])))
		case "int":
			f.WriteString(fmt.Sprintf("    %s int32\n", utils.FormatName(info.keys[i])))
		case "float":
			f.WriteString(fmt.Sprintf("    %s float32\n", utils.FormatName(info.keys[i])))
		case "string":
			f.WriteString(fmt.Sprintf("    %s string\n", utils.FormatName(info.keys[i])))
		}
	}
	f.WriteString("}\n")
	f.WriteString(fmt.Sprintf(`
type %sConfig struct {
    datas map[int32]*%s
}
`, mNameF, mNameF))
	f.WriteString(fmt.Sprintf(`
func (this *%sConfig) Read(dbname string, db *sql.DB) error {
    sql_string := fmt.Sprintf("SELECT * FROM %s.%s", dbname)
        rows, err := db.Query(sql_string)
        if err != nil {
                return err
        }
        return this.scanResult(rows)
}

func (this *%sConfig) GetData(templateId int32) *%s {
    if msg, found := this.datas[templateId]; found {
        return msg
    } else {
        return nil
    }
}
`, mNameF, "`%s`", "`"+mName+"`", mNameF, mNameF))
	// ScanResult
	f.WriteString(fmt.Sprintf(`
func (this *%sConfig) scanResult(res *sql.Rows) error {
    json.Number("").String()
    this.datas = make(map[int32]*%s)
    for res.Next() {
`, mNameF, mNameF))
	for i := 0; i < len(info.keys); i++ {
		if !checkExtraForServer(info.extra[i]) {
			continue
		}
		switch info.types[i] {
		case "int":
			f.WriteString(fmt.Sprintf("        var %s int32\n", info.keys[i]))
		case "float":
			f.WriteString(fmt.Sprintf("        var %s float32\n", info.keys[i]))
		default:
			f.WriteString(fmt.Sprintf("        var %s string\n", info.keys[i]))
		}
	}
	f.WriteString("        scan_err := res.Scan(")
	for i := 0; i < len(info.keys); i++ {
		if !checkExtraForServer(info.extra[i]) {
			continue
		}
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(fmt.Sprintf("&%s", info.keys[i]))
	}
	f.WriteString(")")
	f.WriteString(`
        if scan_err != nil {
            return scan_err
        }

`)
	f.WriteString(fmt.Sprintf("        new%s := &%s{\n", mNameF, mNameF))
	for i := 0; i < len(info.keys); i++ {
		if !checkExtraForServer(info.extra[i]) {
			continue
		}
		if !strings.HasPrefix(info.types[i], "array_") {
			f.WriteString(fmt.Sprintf("            %s: %s,\n", utils.FormatName(info.keys[i]), info.keys[i]))
		}
	}
	f.WriteString("        }\n")
	for i := 0; i < len(info.keys); i++ {
		if !checkExtraForServer(info.extra[i]) {
			continue
		}
		if strings.HasPrefix(info.types[i], "array_") {
			fkey := utils.FormatName(info.keys[i])
			f.WriteString(fmt.Sprintf(`        err%s := json.Unmarshal([]byte(%s), &new%s.%s)
        if err%s != nil {
                return err%s
        }
`, fkey, info.keys[i], mNameF, fkey, fkey, fkey))
		}
	}
	f.WriteString(fmt.Sprintf(`
        if _, found := this.datas[id]; found {
            return fmt.Errorf("duplicate id")
        } else {
            this.datas[id] = new%s
        }
`, mNameF))
	f.WriteString("    }\n")
	f.WriteString(`
    if len(this.datas) == 0 {
        return fmt.Errorf("no data error")
    } else {
        return nil
    }
}
`)
}

func dumpGolangMg() {
	fpath := filepath.Join(cfg.Dir.Out, "ConfigMg.go")
	if pathext.Exists(fpath) {
		os.Remove(fpath)
	}
	f, e := os.Create(fpath)
	defer f.Close()
	utils.CheckError(e)

	f.WriteString(fmt.Sprintf(`//ConfigMg created by tool. Do not modify.
//date: %s
//author: hanxl
`, time.Now().Format("2006-01-02 15:04:05")))
	f.WriteString(`
package config
import (
    "database/sql"
)
`)
	f.WriteString("type ConfigMg struct {\n")
	for _, v := range configList {
		f.WriteString(fmt.Sprintf("    %s %s\n", v, v))
	}
	f.WriteString("}\n")
	f.WriteString(`
func (this *ConfigMg) Init(dbname string, db *sql.DB) error {
    var err error
`)
	for _, v := range configList {
		f.WriteString(fmt.Sprintf(`
    err = this.%s.Read(dbname, db)
    if err != nil {
        return err
    }
`, v))
	}
	f.WriteString("    return nil\n}\n")
}

func cleanupMg() {
	configList = []string{}
}
