package ffjson

import (
	jsoniter "github.com/json-iterator/go"
)

// 使用json-iterator替换go原生的json
var json = jsoniter.ConfigFastest

// 兼容老代码
func JsonUncamel(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte{}, nil
	}
	return json.Marshal(v)
}

// 兼容老代码
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// 兼容老代码
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
