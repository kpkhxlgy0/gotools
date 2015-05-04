package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"crypto/md5"

	"code.google.com/p/gcfg"

	"uuzu.com/hanxl/pathext"
	"uuzu.com/hanxl/utils"
)

type config struct {
	Version struct {
		Cur int
	}
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

	workdir := filepath.Join(pathext.GetExecDir(), "../../..")

	curDir, e := os.Getwd()
	os.Chdir(workdir)
	cmd := exec.Command("svn", "up")
	cmd.Run()
	cmd = exec.Command("svn", "info")
	outBuf, _ := cmd.Output()
	outputs := strings.Split(string(outBuf), "\n")
	os.Chdir(curDir)

	var verLast int
	pattern := regexp.MustCompile(`^Last\sChanged\sRev:\s(.+)`)
	for _, v := range outputs {
		arr := pattern.FindStringSubmatch(v)
		if len(arr) > 0 {
			verLast, _ = strconv.Atoi(strings.Trim(arr[1], "\n\r"))
		}
	}
	println(verLast)

	distdir := filepath.Join(workdir, "tools", "dist")
	if pathext.Exists(distdir) {
		os.RemoveAll(distdir)
		os.MkdirAll(distdir, os.ModePerm)
	} else {
		os.MkdirAll(distdir, os.ModePerm)
	}

	subdirs := [...]string{"res", "src"}
	pattern = regexp.MustCompile(`[\s\t]*[AM](.+)`)
	zipFiles := make(map[string]*zip.Writer)
	for _, subdir := range subdirs {
		fulldir := filepath.Join(workdir, subdir)
		for i := cfg.Version.Cur; i < verLast; i++ {
			os.Chdir(fulldir)
			cmd = exec.Command("svn", "diff", "-r", fmt.Sprintf("%d:%d", i, verLast), "--summarize")
			outBuf, _ := cmd.Output()
			outputs := strings.Split(string(outBuf), "\n")
			os.Chdir(curDir)

			result := []string{}
			for _, v := range outputs {
				arr := pattern.FindStringSubmatch(v)
				if len(arr) > 0 {
					item := bytes.Trim(bytes.Trim([]byte(arr[1]), " "), "\t")
					result = append(result, string(item))
				}
			}

			var zipFile *zip.Writer = nil
			var zipName string
			for _, v := range result {
				fullPath, _ := filepath.Abs(filepath.Join(fulldir, v))
				stat, _ := os.Stat(fullPath)
				if stat.IsDir() {
					continue
				}
				if zipFile == nil {
					verPath := filepath.Join(distdir, fmt.Sprintf("%d_%d", i, verLast))
					if !pathext.Exists(verPath) {
						os.MkdirAll(verPath, os.ModePerm)
					}
					zipName = filepath.Join(verPath, "package")
					zipFile = zipFiles[zipName]
					if zipFile == nil {
						f, e := os.Create(zipName)
						utils.CheckError(e)
						zipFile = zip.NewWriter(f)
						zipFiles[zipName] = zipFile
					}
				}
				// fh := &zip.FileHeader{
				// 	Name:   filepath.Join(subdir, v),
				// 	Method: zip.Deflate,
				// }
				// f, e := zipFile.CreateHeader(fh)
				f, e := zipFile.Create(filepath.Join(subdir, v))
				utils.CheckError(e)
				buf, e := ioutil.ReadFile(fullPath)
				f.Write(buf)
			}
		}
	}

	for zipName, zipFile := range zipFiles {
		zipFile.Close()
		buf, e := ioutil.ReadFile(zipName)
		vPath := filepath.Join(filepath.Dir(zipName), "version")
		f, e := os.Create(vPath)
		utils.CheckError(e)
		f.WriteString(fmt.Sprintf("%x", md5.Sum(buf)))
		f.Close()
	}

	dict := utils.Map{}
	dict.Items = append(dict.Items, utils.Pair{
		Key:   "verLast",
		Value: fmt.Sprintf("%d", verLast),
	})
	e = filepath.Walk(distdir, func(path string, finfo os.FileInfo, err error) error {
		if finfo == nil {
			return err
		} else if !finfo.IsDir() {
			return nil
		} else if distdir == path {
			return nil
		}
		sub := filepath.Base(path)
		pathPackage := filepath.Join(path, "package")
		stat, e := os.Stat(pathPackage)
		utils.CheckError(e)
		dict.Items = append(dict.Items, utils.Pair{
			Key:   "k" + sub,
			Value: int(stat.Size()),
		})
		return nil
	})
	f, e := os.Create(filepath.Join(distdir, "version"))
	defer f.Close()
	utils.CheckError(e)

	f.WriteString("return ")
	utils.DumpLua(reflect.ValueOf(dict), f, 0)

	if runtime.GOOS == "windows" {
		bufio.NewReader(os.Stdin).ReadLine()
	}
}
