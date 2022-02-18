package nio

import (
	"bytes"
	"encoding/binary"
	"github.com/Zoxu0928/task-common/tools"
)

// 提供编解码器，socket交互必须按照此消息结构的规范传递，防止粘包/断包
type Codec interface {

	// 编码，将数据编码为字节
	Encode(msg *Msg) *bytes.Buffer

	// 解码，将数据解码为需要的结构体，返回true代表解码成功，false代表发生断包解码工作还在进行中
	Decode(msg *Msg, reader *bytes.Buffer) bool
}

//-----------------------------------------------------------------------

// 消息体
// flag 与 body 为消息的原始数据，不可改变
// length 与 data 为编码或解码之后动态设置的，由于不同的编解码实现机制不同，所以其具体的值由实现方决定
type Msg struct {

	// 消息类型：占用1字节，最大值255
	flag uint8

	// 消息体
	body interface{}

	// 消息总长度：占用4字节，最大值4294967295
	length uint32

	// 消息体对应的字节数据
	data []byte
}

// Getter、Setter方法
func (msg *Msg) Flag() uint8 {
	return msg.flag
}
func (msg *Msg) Body() interface{} {
	return msg.body
}
func (msg *Msg) Length() uint32 {
	return msg.length
}
func (msg *Msg) Data() []byte {
	return msg.data
}
func (msg *Msg) SetLength(length uint32) {
	msg.length = length
}
func (msg *Msg) SetData(data []byte) {
	msg.data = data
}

// 创建消息体
func CreateMsg(flag uint8, body interface{}) *Msg {
	return &Msg{
		flag: flag,
		body: body,
	}
}

// --------------------------------------------------------------------

// 提供默认的编码解码器

type defaultCodec struct {
	gzip bool
}

func GetDefaultCodec(gzip bool) *defaultCodec {
	return &defaultCodec{gzip: gzip}
}

// 编码
func (c *defaultCodec) Encode(msg *Msg) *bytes.Buffer {

	// 编码结果
	buf := new(bytes.Buffer)

	// 真正数据
	var body_bytes []byte

	// 计算消息长度
	if msg.body == nil {
		msg.length = 1
	} else {
		body_bytes = tools.ToByte(msg.body)
		if c.gzip {
			body_bytes, _ = tools.Gzip(body_bytes)
		}
		msg.length = uint32(len(body_bytes) + 1)
	}

	// 生成数据流
	binary.Write(buf, binary.BigEndian, msg.length)
	binary.Write(buf, binary.BigEndian, msg.flag)
	if body_bytes != nil && len(body_bytes) > 0 {
		buf.Write(body_bytes)
	}
	return buf
}

// 解码
func (c *defaultCodec) Decode(msg *Msg, reader *bytes.Buffer) bool {

	// 说明还没有解码消息头
	if msg.length == 0 {

		// 如果流中数据足够
		if reader.Len() >= 5 {

			// 读取4个字节
			var length uint32
			binary.Read(reader, binary.BigEndian, &length)
			msg.length = length

			// 读取1个字节
			var flag uint8
			binary.Read(reader, binary.BigEndian, &flag)
			msg.flag = flag

			// 流中数据不够，暂时返回
		} else {
			return false
		}
	}

	// 说明这条消息，只有消息头，可以返回了
	if msg.length == 1 {
		return true
	}

	// 还没有解码消息体
	if msg.length > 1 && msg.data == nil {

		// 如果流中数据足够
		if reader.Len() >= int(msg.length-1) {

			// 读取消息体
			body_bytes := make([]byte, msg.length-1)
			reader.Read(body_bytes)
			msg.data = body_bytes

			// 流中数据不够，暂时返回
		} else {
			return false
		}
	}

	if c.gzip {
		msg.data, _ = tools.Gunzip(msg.data)
	}

	// 解码完毕
	return true
}
