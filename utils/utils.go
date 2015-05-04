package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
)

type Pair struct {
	Key   interface{}
	Value interface{}
}

type Map struct {
	Items []Pair
}

func (this *Map) Get(key interface{}) interface{} {
	for _, v := range this.Items {
		if v.Key == key {
			return v.Value
		}
	}
	return nil
}

func Try(fun func(), catch func(interface{}), finaly func()) {
	defer func() {
		if e := recover(); e != nil {
			if catch != nil {
				catch(e)
			}
		}
		if finaly != nil {
			finaly()
		}
	}()
	fun()
}

func CheckError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func FormatName(s string) string {
	arr := strings.Split(s, "_")
	ret := ""
	for _, v := range arr {
		ret += strings.ToUpper(v[0:1]) + strings.ToLower(v[1:])
	}
	return ret
}

func ReadLine(r *bufio.Reader) (string, error) {
	ret := ""
	for {
		line, isPrefix, e := r.ReadLine()
		if e != nil {
			return ret, e
		}
		ret += string(line)
		if !isPrefix {
			return ret, nil
		}
	}
}

func GetKindValue(v reflect.Value) (reflect.Kind, reflect.Value) {
	kind := v.Kind()
	if kind == reflect.Interface {
		tempPtr, ok := v.Interface().(*interface{})
		if ok {
			return reflect.Ptr, reflect.ValueOf(tempPtr)
		}
		tempInt, ok := v.Interface().(int)
		if ok {
			return reflect.Int, reflect.ValueOf(tempInt)
		}
		tempFloat, ok := v.Interface().(float32)
		if ok {
			return reflect.Float32, reflect.ValueOf(tempFloat)
		}
		tempString, ok := v.Interface().(string)
		if ok {
			return reflect.String, reflect.ValueOf(tempString)
		}
		tempSlice, ok := v.Interface().([]interface{})
		if ok {
			return reflect.Slice, reflect.ValueOf(tempSlice)
		}
		tempMap, ok := v.Interface().(map[interface{}]interface{})
		if ok {
			return reflect.Map, reflect.ValueOf(tempMap)
		}
		tempStruct, ok := v.Interface().(Map)
		if ok {
			return reflect.Struct, reflect.ValueOf(tempStruct)
		}
	}
	return kind, v
}

func DumpLua(d reflect.Value, f *os.File, index int) {
	kind, d := GetKindValue(d)
	if kind == reflect.Ptr {
		DumpLua(d.Elem(), f, index)
		return
	}
	sep := ""
	tab := strings.Repeat(" ", 4)
	f.WriteString("{\n")
	if kind == reflect.Array || kind == reflect.Slice {
		for i := 0; i < d.Len(); i++ {
			v := d.Index(i)
			f.WriteString(sep)
			if sep == "" {
				sep = ",\n"
			}
			f.WriteString(strings.Repeat(tab, index+1))
			if v.IsNil() {
				continue
			}
			dumpLuaDefault(v, f, index)
		}
	} else if kind == reflect.Map {
		keys := d.MapKeys()
		for i := 0; i < len(keys); i++ {
			k := keys[i]
			v := d.MapIndex(k)
			dumpLuaKV(k, v, f, index, &sep, tab)
		}
	} else if kind == reflect.Struct {
		items := d.FieldByName("Items")
		for i := 0; i < items.Len(); i++ {
			item := items.Index(i)
			k := item.FieldByName("Key")
			v := item.FieldByName("Value")
			dumpLuaKV(k, v, f, index, &sep, tab)
		}
	}
	f.WriteString("\n" + strings.Repeat(tab, index) + "}")
}

func dumpLuaKV(k, v reflect.Value, f *os.File, index int, sep *string, tab string) {
	f.WriteString(*sep)
	if *sep == "" {
		*sep = ",\n"
	}
	f.WriteString(strings.Repeat(tab, index+1))
	key, k := GetKindValue(k)
	if key == reflect.Int {
		f.WriteString(fmt.Sprintf("[%d]", k.Int()))
	} else {
		f.WriteString(k.String())
	}
	f.WriteString(" = ")
	if v.IsNil() {
		f.WriteString("nil")
	} else {
		dumpLuaDefault(v, f, index)
	}
}

func dumpLuaDefault(v reflect.Value, f *os.File, index int) {
	key, v := GetKindValue(v)
	switch key {
	case reflect.Ptr:
		dumpLuaDefault(v.Elem(), f, index)
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Struct:
		DumpLua(v, f, index+1)
	case reflect.Int:
		f.WriteString(fmt.Sprintf("%d", v.Int()))
	case reflect.Float32:
		f.WriteString(fmt.Sprintf("%.8f", v.Float()))
	default:
		f.WriteString(fmt.Sprintf("\"%s\"", strings.Replace(v.String(), "\n", "\\n", -1)))
	}
}
