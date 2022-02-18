package nio_test

import (
	"bytes"
	"fmt"
	"github.com/Zoxu0928/task-common/nio"
	"testing"
)

func TestSome(t *testing.T) {

	str := `{"code":400,"type":"FAILED_PRECONDITION","message":"InstanceType 'g.s1.small' is out of stock","cause":null}{"code":400,"type":"FAILED_PRECONDITION","message":"InstanceType 'g.s1.small' is out of stock","cause":null}{"code":400,"type":"FAILED_PRECONDITION","message":"InstanceType 'g.s1.small' is out of stock","cause":null}{"code":400,"type":"FAILED_PRECONDITION","message":"InstanceType 'g.s1.small' is out of stock","cause":null}{"code":400,"type":"FAILED_PRECONDITION","message":"InstanceType 'g.s1.small' is out of stock","cause":null}`

	// 消息
	msg := nio.CreateMsg(12, []byte(str))

	// 普通编码器
	normalCodec := nio.GetDefaultCodec(false)

	// 编码
	normalDataEnc := normalCodec.Encode(msg)

	// 解码
	normalMsgDec := &nio.Msg{}
	normalCodec.Decode(normalMsgDec, bytes.NewBuffer(normalDataEnc.Bytes()))

	fmt.Println("编码后：Length:", msg.Length(), "DataLength:", len(normalDataEnc.Bytes()))
	fmt.Println("解码后：Length:", normalMsgDec.Length(), "DataLength:", len(normalMsgDec.Data()), string(normalMsgDec.Data()))

	// 压缩编码器
	gzipCodec := nio.GetDefaultCodec(true)

	// 编码
	gzipDataEnc := gzipCodec.Encode(msg)

	// 解码
	gzipMsgDec := &nio.Msg{}
	gzipCodec.Decode(gzipMsgDec, bytes.NewBuffer(gzipDataEnc.Bytes()))

	fmt.Println("编码后：Length:", msg.Length(), "DataLength:", len(gzipDataEnc.Bytes()))
	fmt.Println("解码后：Length:", gzipMsgDec.Length(), "DataLength:", len(gzipMsgDec.Data()), string(gzipMsgDec.Data()))

}
