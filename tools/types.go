package tools

import (
	"github.com/Zoxu0928/task-common/tools/json/ffjson"
	"reflect"
)

func ToByte(v interface{}) []byte {
	switch t := v.(type) {
	case string:
		return []byte(t)
	case []byte:
		return v.([]byte)
	}
	j, _ := ffjson.JsonUncamel(v)
	return j
}

func byte2bool(bs []byte) bool {
	var c bool
	ffjson.Unmarshal(bs, &c)
	return c
}

func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		if v.Field(i).Kind() == reflect.Struct {
			if t.Field(i).Anonymous {
				struct2Map(data, v.Field(i))
			} else {
				data[LowerFirst(t.Field(i).Name)] = struct2Map(nil, v.Field(i))
			}
		} else {
			data[LowerFirst(t.Field(i).Name)] = v.Field(i).Interface()
		}
	}
	return data
}

func struct2Map(mp map[string]interface{}, obj reflect.Value) map[string]interface{} {
	t := obj.Type()
	v := obj
	var data map[string]interface{}
	if mp == nil {
		data = make(map[string]interface{})
	} else {
		data = mp
	}
	for i := 0; i < t.NumField(); i++ {
		if v.Field(i).Kind() == reflect.Struct {
			if t.Field(i).Anonymous {
				struct2Map(data, v.Field(i))
			} else {
				data[LowerFirst(t.Field(i).Name)] = struct2Map(nil, v.Field(i))
			}
		} else {
			data[LowerFirst(t.Field(i).Name)] = v.Field(i).Interface()
		}
	}
	return data
}
