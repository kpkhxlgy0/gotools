package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/protoParser"
	"uuzu.com/hanxl/utils"
)

func dumpGolang(proto *protoParser.ProtoGo) {
	for _, m := range proto.Messages {
		fname := fmt.Sprintf("%s_db.go", m.Name)
		fpath := filepath.Join(cfg.Dir.Out, fname)
		if pathext.Exists(fpath) {
			os.Remove(fpath)
		}
		f, e := os.Create(fpath)
		defer f.Close()
		utils.CheckError(e)

		f.WriteString(fmt.Sprintf(`//%s created by tool. Do not modify.
//date: %s
//author: hanxl
`, fname, time.Now().Format("2006-01-02 15:04:05")))
		f.WriteString(fmt.Sprintf(`
package db
import (
    "database/sql"
    "encoding/json"
    "fmt"
    "gameserver/proto/%s"
)
import GoogleProto "github.com/golang/protobuf/proto"
`, proto.Package))
		mName := m.Name
		mNameF := utils.FormatName(mName)
		mNameTotal := fmt.Sprintf("%s.%s", proto.Package, mNameF)
		f.WriteString(fmt.Sprintf(`
type %sDB struct {
    one_key        uint64
    sec_key        uint64
    update         *%s
    querys         []*%s
    op             int
    update_val_map map[string]interface{}
}
`, mNameF, mNameTotal, mNameTotal))
		f.WriteString(fmt.Sprintf(`
func New%sDB(one_key, sec_key uint64, update *%s) *%sDB {
    return &%sDB{
        one_key:        one_key,
        sec_key:        sec_key,
        update:         update,
        update_val_map: make(map[string]interface{}),
    }
}
`, mNameF, mNameTotal, mNameF, mNameF))
		f.WriteString(fmt.Sprintf(`
func (this *%sDB) GetOpClass() int {
    return this.op
}

func (this *%sDB) GetOneKey() uint64 {
    return this.one_key
}
`, mNameF, mNameF))
		// ScanResult
		f.WriteString(fmt.Sprintf(`
func (this *%sDB) ScanResult(res *sql.Rows) error {
    for res.Next() {
        var one_key uint64
        var sec_key uint64
`, mNameF))
		for _, v := range m.Value {
			if v.IsArr {
				f.WriteString(fmt.Sprintf("        var %s string\n", v.Name))
			} else {
				f.WriteString(fmt.Sprintf("        var %s %s\n", v.Name, v.TypeStr))
			}
		}
		f.WriteString("        scan_err := res.Scan(&one_key, &sec_key")
		for _, v := range m.Value {
			f.WriteString(fmt.Sprintf(", &%s", v.Name))
		}
		f.WriteString(")")
		f.WriteString(`
        if scan_err != nil {
            return scan_err
        }
`)
		mNameNew := fmt.Sprintf("new%s", mNameF)
		f.WriteString(fmt.Sprintf(`
        %s := &%s{
            OneKey: GoogleProto.Uint64(0),
            SecKey: GoogleProto.Uint64(0),
`, mNameNew, mNameTotal))
		for _, v := range m.Value {
			f.WriteString(fmt.Sprintf("            %s: ", utils.FormatName(v.Name)))
			if v.IsArr {
				f.WriteString(fmt.Sprintf("make([]%s, 10),\n", v.TypeStr))
			} else if v.TypeStr == "string" {
				f.WriteString("GoogleProto.String(\"\"),\n")
			} else {
				f.WriteString(fmt.Sprintf("GoogleProto.%s(0),\n", utils.FormatName(v.TypeStr)))
			}
		}
		f.WriteString("        }")
		f.WriteString(fmt.Sprintf(`
        *%s.OneKey = one_key
        *%s.SecKey = sec_key
`, mNameNew, mNameNew))
		for _, v := range m.Value {
			if v.IsArr {
				fkey := utils.FormatName(v.Name)
				f.WriteString(fmt.Sprintf(`        err%s := json.Unmarshal([]byte(%s), &%s.%s)
        if err%s != nil {
            return err%s
        }
`, fkey, v.Name, mNameNew, fkey, fkey, fkey))
			} else {
				f.WriteString(fmt.Sprintf("        *%s.%s = %s\n", mNameNew, utils.FormatName(v.Name), v.Name))
			}
		}
		f.WriteString(fmt.Sprintf("        this.querys = append(this.querys, %s)\n", mNameNew))
		f.WriteString("    }")
		f.WriteString(`
    if len(this.querys) == 0 {
        return fmt.Errorf("no data error")
    } else {
        return nil
    }
}
`)
		f.WriteString(fmt.Sprintf(`
func (this *%sDB) SelectAll() {
    this.op = OP_SELECT_ALL
}
`, mNameF))
		f.WriteString(fmt.Sprintf("func (this *%sDB) UpdateAll() {\n", mNameF))
		f.WriteString("    this.op = OP_UPDATE\n")
		for _, v := range m.Value {
			f.WriteString(fmt.Sprintf("    this.Update%s()\n", utils.FormatName(v.Name)))
		}
		f.WriteString("}\n")
		f.WriteString(fmt.Sprintf(`
func (this *%sDB) Insert() {
    this.op = OP_INSERT
}

func (this *%sDB) Delete() {
    this.op = OP_DEL
}
`, mNameF, mNameF))
		for _, v := range m.Value {
			f.WriteString(fmt.Sprintf(`
func (this *%sDB) Update%s() {
    this.op = OP_UPDATE
    this.update_val_map["%s"] = this.update.Get%s()
}
`, mNameF, utils.FormatName(v.Name), v.Name, utils.FormatName(v.Name)))
		}
		f.WriteString(fmt.Sprintf(`
func (this *%sDB) FormatSQL(DbName string) (string, error) {
    if this.one_key == 0 {
        return "", fmt.Errorf("one key is 0")
    }
    switch this.op {
    case OP_SELECT_ALL:
        return this.formatSelectAll(DbName), nil
    case OP_DEL:
        return this.formatDel(DbName), nil
    case OP_INSERT:
        if this.sec_key == 0 {
            return "", fmt.Errorf("insert op second key is 0")
        }
        return this.formatInsert(DbName), nil
    case OP_UPDATE:
        if this.sec_key == 0 {
            return "", fmt.Errorf("insert op second key is 0")
        }
        return this.formatUpdate(DbName), nil
    }
    return "", fmt.Errorf("not define db op")
}

func (this *%sDB) Message() []*%s {
    return this.querys
}
`, mNameF, mNameF, mNameTotal))
		f.WriteString(fmt.Sprintf(`
func (this *%sDB) formatSelectAll(DbName string) string {
    if this.sec_key != 0 {
        return fmt.Sprintf("SELECT * FROM %s.%s WHERE %s= %%d AND %s= %%d", DbName, this.one_key, this.sec_key)
    } else {
        return fmt.Sprintf("SELECT * FROM %s.%s WHERE %s= %%d", DbName, this.one_key)
    }
}

func (this *%sDB) formatDel(DbName string) string {
    return fmt.Sprintf("DELETE  FROM %s.%s WHERE %s= %%d AND %s= %%d", DbName, this.one_key, this.sec_key)
}
`, mNameF, "`%s`", "`"+mName+"`", "`one_key`", "`sec_key`", "`%s`", "`"+mName+"`", "`one_key`", mNameF, "`%s`", "`"+mName+"`", "`one_key`", "`sec_key`"))
		f.WriteString(fmt.Sprintf("func (this *%sDB) formatInsert(DbName string) string {\n", mNameF))
		for _, v := range m.Value {
			if v.IsArr {
				f.WriteString(fmt.Sprintf("    %s, _ := json.Marshal(this.update.Get%s())\n", v.Name, utils.FormatName(v.Name)))
			}
		}
		f.WriteString(fmt.Sprintf("    return fmt.Sprintf(\"INSERT INTO `%%s`.`%s` (`one_key`, `sec_key`", mName))
		for _, v := range m.Value {
			f.WriteString(fmt.Sprintf(", `%s`", v.Name))
		}
		f.WriteString(") VALUES(%v,%v")
		for _, v := range m.Value {
			if v.IsArr || v.TypeStr == "string" {
				f.WriteString(",'%v'")
			} else {
				f.WriteString(",%v")
			}
		}
		f.WriteString(")\",")
		f.WriteString(`
        DbName,
        this.one_key,
        this.sec_key,
`)
		i := 0
		for _, v := range m.Value {
			if i > 0 {
				f.WriteString(",\n")
			}
			f.WriteString("        ")
			if v.IsArr {
				f.WriteString(fmt.Sprintf("string(%s)", v.Name))
			} else {
				f.WriteString(fmt.Sprintf("this.update.Get%s()", utils.FormatName(v.Name)))
			}
			i++
		}
		f.WriteString(")\n}\n")
		f.WriteString(fmt.Sprintf(`
func (this *%sDB) formatUpdate(DbName string) string {
    var selct_sq = ""
    var leng = len(this.update_val_map)
    var i int = 0
    for col, val := range this.update_val_map {
        var _sq string
        switch val.(type) {
        case []byte, []int8, []int32, []uint32, []int16, []uint16, []uint64, []int64, []string:
            _data, _ := json.Marshal(val)
            _sq = fmt.Sprintf("%s = '%%v' ", col, string(_data))
        case string:
            _sq = fmt.Sprintf("%s = '%%v' ", col, val)
        default:
            _sq = fmt.Sprintf("%s = %%v ", col, val)
        }
        if (i + 1) != leng {
            _sq = fmt.Sprintf("%%s,", _sq)
        }
        i += 1
        selct_sq = selct_sq + _sq
    }
    return fmt.Sprintf("UPDATE %s.%s  SET %%s WHERE %s= %%d AND %s= %%d ",
        DbName,
        selct_sq,
        this.one_key,
        this.sec_key)
}
`, mNameF, "`%s`", "`%s`", "`%s`", "`%s`", "`"+mName+"`", "`one_key`", "`sec_key`"))
	}
}
