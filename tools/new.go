package tools

// 初始化指针类型值

func NewString(v string) *string {
	return &v
}

func NewBool(v bool) *bool {
	return &v
}

func NewInt(v int) *int {
	return &v
}

func NewInt64(v int) *int64 {
	i := int64(v)
	return &i
}
