package tools

// DeleteStringSlice 删除字符串切片内相同的字符串
func DeleteStringSlice(origin []string, str string) []string {
	i := 0
	for _, v := range origin {
		if str != v {
			origin[i] = v
			i++
		}
	}
	return origin[:i]
}
