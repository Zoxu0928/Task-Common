package nio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/Zoxu0928/task-common/e"
	"github.com/Zoxu0928/task-common/logger"

	"net"
	"sync/atomic"
	"time"
)

var session_id int64 = 0

// Socket客户端
type SocketConnector struct {
	socket
	session *Session //连接会话
}

// 创建Socket客户端
func CreateConnector(ip, port string, handler *Handler, codec Codec, ordered bool) (*SocketConnector, error) {
	if handler == nil {
		return nil, errors.New(ip + ":" + port + " Create socket connector failed. Handler is nil.")
	}
	connector := &SocketConnector{}
	connector.addr = ip + ":" + port
	connector.handler = handler
	connector.ordered = ordered
	connector.handler.codec = codec
	if codec == nil {
		connector.handler.codec = GetDefaultCodec(false)
	}

	return connector, nil
}

// 开始连接
func (socket *SocketConnector) Connect() error {

	logger.Info("[client] Connect to server %s ...", socket.addr)

	// 连接
	connection, err := net.DialTimeout("tcp", socket.addr, 10*time.Second)
	if err != nil {
		return err
	}

	// 创建一个Session
	session := &Session{}
	session.id = atomic.AddInt64(&session_id, 1)
	session.isClient = true
	session.codec = socket.handler.codec
	session.connector = socket
	session.connection = connection

	socket.session = session

	// 等待并读取服务端发送过来的会话ID
	if sessionId, sync_err := socket.syncSessionId(); sync_err != nil {
		logger.Error("ERROR %s [client] Sync Session Id failed. Cause of %s\n", socket.addr, sync_err.Error())
		connection.Close()
		return sync_err
	} else {
		session.id = sessionId
	}

	logger.Info("[client] [%d] Connect to server %s success. Local %s", session.id, socket.addr, connection.LocalAddr())

	// 同步回调，否则有可能出现数据不一致
	if socket.handler.OnSessionConnected != nil {
		func() {
			defer e.OnError("")
			socket.handler.OnSessionConnected(session)
		}()
	}

	// 异步处理网络流
	go socket.handler.readIo(&socket.socket, session)

	return nil
}

// 向服务端发送数据
func (socket *SocketConnector) WriteMsg(msg *Msg) error {
	if err := socket.session.WriteMsg(msg); err != nil {
		return err
	}
	return nil
}

// 向服务端发送数据
func (socket *SocketConnector) Write(flag uint8, data interface{}) error {
	if err := socket.session.Write(flag, data); err != nil {
		return err
	}
	return nil
}

// 关闭连接
func (socket *SocketConnector) Close() {

	logger.Info("[Close Event] [client] [%d] Close socket connection %s.", socket.session.id, socket.addr)

	socket.closed = true

	// 异步回调
	if socket.handler.OnSessionClosed != nil {
		go func() {
			defer e.OnError("")
			socket.handler.OnSessionClosed(socket.session)
		}()
	}

	// 关闭socket
	for i := 1; i <= 3; i++ {
		if err := socket.session.connection.Close(); err != nil {
			logger.Error("[client] [%d] Close socket connection %s error. %s", socket.session.id, socket.addr, err.Error())
			time.Sleep(1 * time.Second)
		} else {
			return
		}
	}
}

// 接收SessionId
func (socket *SocketConnector) syncSessionId() (int64, error) {
	defer e.OnError("Sync Session Id")
	var sessionId int64
	first_data := bytes.NewBuffer([]byte{})
	first_data_len := 8
	for {
		buf := make([]byte, first_data_len)
		len, io_err := socket.session.connection.Read(buf)
		if io_err != nil {
			return 0, io_err
		}
		if len > 0 {
			first_data.Write(buf[0:len])
		}
		first_data_len = first_data_len - len
		if first_data_len == 0 {
			break
		}
	}
	binary.Read(first_data, binary.BigEndian, &sessionId)
	return sessionId, nil
}
