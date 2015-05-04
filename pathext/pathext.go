package pathext

import (
	"os"
	"path/filepath"
	"runtime"
)

func GetExecPath() string {
	s := os.Args[0]
	path, _ := filepath.Abs(s)
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		base := filepath.Base(path)
		path = filepath.Join(path, "..", "..", base)
	}
	return path
}

func GetExecDir() string {
	return filepath.Dir(GetExecPath())
}

func GetExecFile() string {
	return filepath.Base(GetExecPath())
}

func BaseWithoutExt(path string) string {
	return path[:len(path)-len(filepath.Ext(path))]
}

func Exists(file string) bool {
	_, err := os.Stat(file)
	return err == nil || os.IsExist(err)
}

func Walk(root string, walkFn filepath.WalkFunc) error {
	paths := new([]string)
	finfos := new([]os.FileInfo)
	errs := new([]error)
	e := walk(root, paths, finfos, errs)
	if e != nil {
		return e
	}
	for i := 0; i < len(*paths); i++ {
		e = walkFn((*paths)[i], (*finfos)[i], (*errs)[i])
		if e != nil {
			return e
		}
	}
	return nil
}

func walk(root string, paths *[]string, finfos *[]os.FileInfo, errs *[]error) error {
	e := filepath.Walk(root, func(path string, finfo os.FileInfo, err error) error {
		if finfo == nil {
			return err
		}
		if root == path {
			return nil
		}
		if finfo.IsDir() {
			return walk(path, paths, finfos, errs)
		} else {
			*paths = append(*paths, path)
			*finfos = append(*finfos, finfo)
			*errs = append(*errs, err)
		}
		return nil
	})
	return e
}
