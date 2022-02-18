package tools

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// GenerateTemplateReplace 对模板字符串进行替换
// 例如: func("/v1/api/{id}", {id: 123}) => ("/v1/api/123", nil)
func GenerateTemplateReplace(str string, paths map[string]interface{}) (string, error) {
	if strings.Index(str, "{") > -1 {
		r := regexp.MustCompile("{([^}]+)}")
		rStr := r.FindString(str)
		if rStr == "" {
			return str, errors.New("无法匹配变量名")
		}
		end := len(rStr)
		paramsName := rStr[1 : end-1]

		rValue := reflect.ValueOf(paths).MapIndex(reflect.ValueOf(paramsName))
		if !rValue.IsValid() {
			return str, errors.New("无法解析 Path:" + paramsName)
		}

		nStr := rValue.Interface()
		switch nStr.(type) {
		case string:
			str = strings.Replace(str, rStr, nStr.(string), -1)
		case int32:
			str = strings.Replace(str, rStr, strconv.FormatInt(int64(nStr.(int32)), 10), -1)
		case int64:
			str = strings.Replace(str, rStr, strconv.FormatInt(nStr.(int64), 10), -1)
		case int:
			str = strings.Replace(str, rStr, strconv.Itoa(nStr.(int)), -1)
		default:
			return str, errors.New("不支持此参数类型 Path:" + paramsName)
		}
		return GenerateTemplateReplace(str, paths)
	}
	return str, nil
}
