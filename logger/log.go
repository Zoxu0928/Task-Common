package logger

// 用途：打印日志

// 如果没有手工配置，使用默认配置
// 默认情况下日志级别为DEBUG，日志输出到控制台

import (
	"fmt"
	"strings"
)

// 声明类
type Conf struct {
	Path       string `toml:"path"`
	Name       string `toml:"file_name"`
	Level      string `toml:"level"`
	MaxHistory int32  `toml:"max_history"`
	Rolling    string `toml:"rolling"`
	Size       int32  `toml:"size"`
}

// 初始化
func InitLogger(cfg *Conf) {
	// 对没有配置的信息初始化默认值
	if cfg.Path == "" {
		setConsole(true)
		setLevel(_DEBUG)
	} else {
		setConsole(false)
		if strings.EqualFold(cfg.Level, "info") {
			setLevel(_INFO)
		} else if strings.EqualFold(cfg.Level, "debug") {
			setLevel(_DEBUG)
		} else {
			setLevel(_DEBUG)
		}
		switch cfg.Rolling {
		case "size":
			if cfg.Size <= 0 {
				cfg.Size = 500
			}
			setRollingFile(cfg.Path, cfg.Name, int32(cfg.MaxHistory), int64(cfg.Size), mb)
		default:
			setRollingDaily(cfg.Path, cfg.Name, int32(cfg.MaxHistory))
		}
	}

	if cfg.Path == "" {
		fmt.Printf("********** Logger init. target -> %s **********\n", "console")
	} else {
		fmt.Printf("********** Logger init. target -> %s **********\n", cfg.Path+"/"+cfg.Name)
	}
}

func Debug(text string, format ...interface{}) {
	defaultlog.debug(text, format...)
}
func Info(text string, format ...interface{}) {
	defaultlog.info(text, format...)
}
func Warn(text string, format ...interface{}) {
	defaultlog.warn(text, format...)
}
func Error(text string, format ...interface{}) {
	defaultlog.error(text, format...)
}
