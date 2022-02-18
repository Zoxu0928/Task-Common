package nio

import (
	"errors"
	"net"
)

// 客户端或服务端保持的会话
type Session struct {
	id         int64                  // 唯一id
	codec      Codec                  // 消息编解码器
	connection net.Conn               // 保存与对方的连接
	attribute  map[string]interface{} // 缓存一些属性，不会同步到服务端或客户端，只在本侧有效
	isClient   bool                   // true代表客户端，false代表服务端
	accecptor  *SocketAccecptor       // 服务端会话所属的acceptor
	connector  *SocketConnector       // 客户端会话所属的connector
}

// 获得会话id
func (session *Session) GetSessionId() int64 {
	return session.id
}

// 向会话中增加一个属性
func (session *Session) SetAttribute(k string, v interface{}) {
	if session.attribute == nil {
		session.attribute = make(map[string]interface{})
	}
	session.attribute[k] = v
}

// 从会话中获得一个属性
func (session *Session) GetAttribute(k string) interface{} {
	if session.attribute == nil {
		return nil
	}
	return session.attribute[k]
}

// 向会话中写入数据
func (session *Session) WriteMsg(msg *Msg) error {
	if session.connection == nil {
		return errors.New("not connected")
	}
	if _, err := session.connection.Write(session.codec.Encode(msg).Bytes()); err != nil {
		return err
	}
	return nil
}

// 向会话中写入数据
func (session *Session) Write(flag uint8, data interface{}) error {
	if session.connection == nil {
		return errors.New("not connected")
	}
	d := session.codec.Encode(CreateMsg(flag, data)).Bytes()
	if _, err := session.connection.Write(d); err != nil {
		return err
	}
	return nil
}

func (session *Session) GetLocalAddr() net.Addr {
	if session.connection != nil {
		return session.connection.LocalAddr()
	}
	return nil
}
func (session *Session) GetRemoteAddr() net.Addr {
	if session.connection != nil {
		return session.connection.RemoteAddr()
	}
	return nil
}

// 关闭会话
func (session *Session) Close() {

	// 作为客户端，直接关闭
	if session.isClient {
		session.connector.Close()

		// 作为服务端，关闭指定session
	} else {
		session.accecptor.CloseSession(session)
	}
}
