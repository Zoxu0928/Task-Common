package tools

import (
	"fmt"
	"github.com/Zoxu0928/task-common/tools/json/ffjson"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// 判断字符串是否以指定字符串开头
func StartWith(s, start string) bool {
	return strings.Index(s, start) == 0
}

// 判断字符串是否以指定字符串结尾
func EndWith(s, end string) bool {
	pattern := regexp.MustCompile(end + `$`)
	return pattern.Match([]byte(s))
}

// 判断字符串是否匹配正则表达式
func MatchReg(s, reg string) bool {
	pattern := regexp.MustCompile(reg)
	return pattern.Match([]byte(s))
}

// 判断字符串item是否在指定列表中
func ContainsString(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}

// 判断字符串item是否在指定列表中
// noMatch如果为true，则忽略大小写
func ContainsString2(list []string, item string, noMatch bool) bool {
	if noMatch {
		item = strings.ToLower(item)
		for _, s := range list {
			if strings.ToLower(s) == item {
				return true
			}
		}
		return false
	} else {
		for _, s := range list {
			if s == item {
				return true
			}
		}
		return false
	}
}

// 将首字母转为小写
func LowerFirst(str string) string {
	if str == "" {
		return str
	} else {
		pattern := regexp.MustCompile("(^[A-Z])")
		char := pattern.FindString(str)
		if char != "" {
			return strings.ToLower(char) + strings.TrimLeft(str, char)
		}
	}
	return str
}

// 匹配正则表达式，进行字符串替换
func Replace(str string, reg string, newStr string) string {
	pattern, err := regexp.Compile(reg)
	if err != nil {
		return str
	}
	return pattern.ReplaceAllString(str, newStr)
}

// 对路径进行匹配
func MatchPath(str string, reg string) bool {
	reg = "^" + strings.Replace(reg, "*", "[0-9a-zA-Z_]*", -1)
	pattern, err := regexp.Compile(reg)
	if err != nil {
		return false
	}
	return pattern.Match([]byte(str))
}

// 将参数转换成Json
func ToJson(v interface{}) string {
	j, _ := ffjson.JsonUncamel(v)
	return string(j)
}
func Marshal(v interface{}) ([]byte, error) {
	return ffjson.Marshal(v)
}
func Unmarshal(data []byte, v interface{}) error {
	return ffjson.Unmarshal(data, v)
}

// 将值转换为字符串
func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	if t := reflect.ValueOf(v); t.Kind() == reflect.Ptr && t.IsNil() {
		return ""
	}
	return fmt.Sprint(v)
}

// 将字符串数组转成int数组
func ToIntArray(s []string) ([]int, error) {
	if s == nil {
		return nil, nil
	}
	vs := make([]int, len(s))
	for i, v := range s {
		if val, err := ToInt(v); err != nil {
			return nil, err
		} else {
			vs[i] = val
		}
	}
	return vs, nil
}

// 将字符串转成int
func ToInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	v, err := strconv.Atoi(s)
	return v, err
}

// 将字符串转成int32
func ToInt32(s string) (int32, error) {
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 32)
	return int32(v), err
}

// 将字符串转成int64
func ToInt64(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	return int64(v), err
}

// 将字符串转成float32
func ToFloat32(s string) (float32, error) {
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 32)
	return float32(v), err
}

// 将字符串数组转成bool数组
func ToBoolArray(s []string) ([]bool, error) {
	if s == nil {
		return nil, nil
	}
	vs := make([]bool, len(s))
	for i, v := range s {
		if val, err := ToBool(v); err != nil {
			return nil, err
		} else {
			vs[i] = val
		}
	}
	return vs, nil
}

// 将字符串转成bool
func ToBool(s string) (bool, error) {
	if s == "" {
		return false, nil
	}
	v, err := strconv.ParseBool(s)
	return v, err
}

// 转换为驼峰形式
func ToCamel(str string) string {
	parts := strings.Split(str, "_")
	name := ""
	for _, v := range parts {
		name = name + strings.Title(v)
	}
	return name
}

// 判断字符串指针是否为空
func IsBlank(str *string) bool {
	return str == nil || *str == ""
}

// 字符串长度
func Length(str string) int {
	return len([]rune(str))
}
