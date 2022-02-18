package tools

import (
	"encoding/xml"
	etree "github.com/Zoxu0928/task-common/tools/xml"
)

// 将对象解析为xml，并且key的首字母转为小写
func XmlUncamel(v interface{}) ([]byte, error) {
	x, _ := xml.Marshal(v)
	d := etree.NewDocument()
	d.ReadFromBytes(x)
	b, err := d.WriteToBytes()
	return b, err
}
