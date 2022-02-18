package tools

import (
	"github.com/Zoxu0928/task-common/logger"
	"io/ioutil"
	"os"
)

// 创建目录
func MkDir(dir string, perm os.FileMode) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, perm); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

// 列出目录下内容
func ListDir(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logger.Error("list dir failed. err=%v", err)
	}
	list := make([]string, len(files))
	for i, v := range files {
		list[i] = v.Name()
	}
	return list
}
