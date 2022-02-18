package env

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 默认情况：配置文件所在路径
var ConfigPath = "cfg"

// 程序运行带文件名的全路径，如 D:/openapi/api-vm.exe
var ProgramPath = ""

func init() {

	// 设置程序运行路径
	file, _ := exec.LookPath(os.Args[0])
	ApplicationPath, _ := filepath.Abs(file)
	ProgramPath = strings.Replace(ApplicationPath, "\\", "/", -1)
}
