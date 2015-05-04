package main

import (
	"os"
	"path/filepath"
	"reflect"

	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/protoParser"
	"uuzu.com/hanxl/utils"
)

func dumpEnum(proto *protoParser.Proto) {
	fpath := filepath.Join(cfg.Dir.Out, "common.lua")
	if pathext.Exists(fpath) {
		os.Remove(fpath)
	}
	f, e := os.Create(fpath)
	defer f.Close()
	utils.CheckError(e)

	f.WriteString("return ")
	utils.DumpLua(reflect.ValueOf(proto), f, 0)
}

func dumpOpcode(opcode *protoParser.Opcode) {
	fpath := filepath.Join(cfg.Dir.Out, "proto.lua")
	if pathext.Exists(fpath) {
		os.Remove(fpath)
	}
	f, e := os.Create(fpath)
	defer f.Close()
	utils.CheckError(e)

	f.WriteString("return ")
	utils.DumpLua(reflect.ValueOf(opcode), f, 0)
}
