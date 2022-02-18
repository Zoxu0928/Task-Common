package tools

import (
	"github.com/Zoxu0928/task-common/tools/json/ffjson"
	"reflect"
)

func GetFieldsByTag(obj interface{}, tag string) []string {
	if obj == nil || tag == "" {
		return nil
	}

	typeValue := reflect.TypeOf(obj)
	if typeValue.Kind() == reflect.Ptr {
		typeValue = reflect.TypeOf(obj).Elem()
	}
	if typeValue.Kind() != reflect.Struct {
		return nil
	}

	fields := make([]string, 0)
	for index := 0; index < typeValue.NumField(); index++ {
		if tagValue := typeValue.Field(index).Tag.Get(tag); tagValue != "" {
			fields = append(fields, tagValue)
		}
	}
	return fields
}

// LoadFromMap 解析 interface 值到结构
func LoadFromMap(target, source interface{}) error {
	data, err := ffjson.Marshal(source)
	if err == nil {
		err = ffjson.Unmarshal(data, target)
	}
	return err
}
