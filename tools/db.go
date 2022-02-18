package tools

import (
	"bytes"
)

// 生成len个占位符
// 例如：symbol=?   split=,   len=3
// 返回 ?,?,?
func CreateSymbol(symbol, split string, len int) string {
	str := new(bytes.Buffer)
	for i := 0; i < len; i++ {
		if i > 0 {
			str.WriteString(split)
		}
		str.WriteString(symbol)
	}
	return str.String()
}
