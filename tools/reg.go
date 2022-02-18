package tools

import (
	"regexp"
	"strings"
)

func CheckRegMatch(strToBeChecked, reg string, length int) bool {
	if strToBeChecked == "" || len(strToBeChecked) > length {
		return false
	}
	pattern := regexp.MustCompile(reg)
	if !pattern.Match([]byte(strToBeChecked)) {
		return false
	}

	return true
}

// 将字符串正则化
func ToRegStr(str string, fix bool) string {

	// 以符号 [ 开头后面的视为字母
	char := false

	// 转义
	escape := false

	// 缓存结果
	buf := strings.Builder{}

	if fix && strings.HasPrefix(str, "^") == false {
		buf.WriteString("^")
	}

	// 处理每一个字符
	for _, n := range str {

		l := string(n)

		// 转义后面的原样输出，然后结束转义
		if escape {
			buf.WriteString(l)
			escape = false
			continue
		}

		if l == `\` && !escape { //转义开始
			escape = true
		} else if l == `[` { //字母开始
			char = true
		} else if char && l == `]` { //字母结束
			char = false
		}

		// 将 * 转换为 .*
		if l == `*` && !escape && !char {
			buf.WriteString(".")
			buf.WriteString(l)
			continue
		}

		// 将 ? 转换为 .
		if l == `?` && !escape && !char {
			buf.WriteString(".")
			continue
		}

		buf.WriteString(l)
	}

	if fix && strings.HasSuffix(str, "$") == false {
		buf.WriteString("$")
	}

	return buf.String()
}
