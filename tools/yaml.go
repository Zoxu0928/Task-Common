package tools

import (
	"github.com/Zoxu0928/task-common/global/env"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// 加载配置文件到结构体中
func LoadYaml(yamlFile string, target interface{}) error {

	// 取配置文件
	content, err := ioutil.ReadFile(env.ConfigPath + "/" + yamlFile)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(content, target); err != nil {
		return err
	}

	return nil
}
