package nio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/Zoxu0928/task-common/e"
	"github.com/Zoxu0928/task-common/logger"

	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// 客户端与服务端共用的结构
type socket struct {
	addr    string   //对于服务端:代表监听的地址，对于客户端:代表要连接的地址
	handler *Handler //业务处理器，由使用方自行实现，必须设置
	closed  bool     //服务端或客户端是否已关闭
	ordered bool     //true消息处理顺序请求，false并发请求
}

// Socket服务端
type SocketAccecptor struct {
	socket
	lock      sync.Mutex
	sessions  map[int64]*Session //保存所有连接到本服务端的会话
	accecptor net.Listener
}

// 创建一个Socket服务端
func CreateAccecptor(ip, port string, handler *Handler, codec Codec, ordered bool) (*SocketAccecptor, error) {
	if handler == nil {
		return nil, errors.New(ip + ":" + port + " Create socket accecptor failed. Handler is nil.")
	}
	accecptor := &SocketAccecptor{}
	accecptor.addr = ip + ":" + port
	accecptor.handler = handler
	accecptor.ordered = ordered
	accecptor.sessions = make(map[int64]*Session)
	accecptor.handler.codec = codec
	if codec == nil {
		accecptor.handler.codec = GetDefaultCodec(false)
	}
	return accecptor, nil
}

// 打开Socket服务端
func (socket *SocketAccecptor) Open() error {

	logger.Info("[server] Open socket %s accecptor and listening...", socket.addr)

	// 打开socket
	accepter, err := net.Listen("tcp", socket.addr)
	if err != nil {
		logger.Error("[server] Open socket %s accecptor failed. %s", socket.addr, err.Error())
		return err
	}
	socket.accecptor = accepter

	logger.Info("[server] Open socket %s accecptor success.", socket.addr)

	// 异步处理连接请求
	go func() {
		defer e.OnError("")
		socket.handlerConnection()
	}()

	return nil
}

// 处理连接
func (socket *SocketAccecptor) handlerConnection() {

	// sessionid，每个会话唯一，递增
	var session_id int64 = 0

	// 待等连接进入
	for {

		if socket.closed {
			logger.Info("[server] Socket accecptor %s is closed. Normal finished.", socket.addr)
			return
		}
		connection, err := socket.accecptor.Accept()
		if err != nil {
			logger.Error("[server] [%s] Open socket connection failed. %s", socket.addr, err.Error())
			break
		}

		// 创建一个Session
		session := &Session{}
		session.id = atomic.AddInt64(&session_id, 1)
		session.isClient = false
		session.codec = socket.handler.codec
		session.accecptor = socket
		session.connection = connection

		// 向客户端通知本次会话的SessionID
		if send_err := socket.sendSessionId(connection, session.id); send_err != nil {
			session.id = atomic.AddInt64(&session_id, -1)
			connection.Close()
			continue
		}

		logger.Info("[server] [%s] [%d] Accecpt connection from remote %s", socket.addr, session.id, connection.RemoteAddr())

		// 保存会话
		socket.addSession(session)

		// 同步回调，否则有可能出现数据不一致
		if socket.handler.OnSessionConnected != nil {
			func() {
				defer e.OnError("")
				socket.handler.OnSessionConnected(session)
			}()
		}

		// 异步处理网络流
		go socket.handler.readIo(&socket.socket, session)
	}
}

// 关闭socket
func (socket *SocketAccecptor) Close() {

	socket.lock.Lock()
	defer socket.lock.Unlock()

	logger.Info("[Close Event] [server] %s Close socket accecptor.", socket.addr)

	socket.closed = true

	// 异步回调
	if socket.handler.OnSessionClosed != nil {
		for _, session := range socket.sessions {
			go func() {
				defer e.OnError("")
				socket.handler.OnSessionClosed(session)
			}()
		}
	}

	// 关闭socket
	for i := 1; i <= 3; i++ {
		if err := socket.accecptor.Close(); err != nil {
			logger.Error("[server] %s Close socket accecptor error. %s", socket.addr, err.Error())
			time.Sleep(1 * time.Second)
		} else {
			return
		}
	}
}

// 关闭一个会话
func (socket *SocketAccecptor) CloseSession(session *Session) {

	// 清除会话
	socket.removeSession(session.id)

	// 异步回调
	if socket.handler.OnSessionClosed != nil {
		go func() {
			defer e.OnError("")
			socket.handler.OnSessionClosed(session)
		}()
	}

	// 关闭socket
	for i := 1; i <= 3; i++ {
		if err := session.connection.Close(); err != nil {
			if strings.Contains(err.Error(), "use of closed network") {
				return
			}
			logger.Error("[client] [%d] Close socket connection %s error. %s", session.id, socket.addr, err.Error())
			time.Sleep(1 * time.Second)
		} else {
			return
		}
	}
}

// 增加一个session
func (socket *SocketAccecptor) addSession(session *Session) {
	socket.lock.Lock()
	defer socket.lock.Unlock()
	socket.sessions[session.id] = session
}

// 删除一个session
func (socket *SocketAccecptor) removeSession(session_id int64) {
	socket.lock.Lock()
	defer socket.lock.Unlock()
	delete(socket.sessions, session_id)
}

// 向对方发送SessionId
func (socket *SocketAccecptor) sendSessionId(connection net.Conn, sessionId int64) error {
	defer e.OnError("Send Session Id")
	time.Sleep(5 * time.Millisecond)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, sessionId)
	if _, io_err := connection.Write(buf.Bytes()); io_err != nil {
		return io_err
	}
	return nil
}

// 会话数量
func (socket *SocketAccecptor) SessionCount() int {
	return len(socket.sessions)
}
