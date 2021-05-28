package common

import (
	"log"
	"os"
	"path/filepath"
)

var workDirOverride string = ""

func SetWorkDir(dir string) {
	workDirOverride = dir
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func JoinCurrentPath(filename string) string {
	ret := workDirOverride
	var err error
	if ret == "" {
		ret, err = os.Executable()
		if err != nil {
			log.Fatalln(err.Error())
		}
	}

	return filepath.Join(filepath.Dir(ret), filename)
}
