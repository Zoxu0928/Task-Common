package nio

import (
	"bytes"
	"github.com/Zoxu0928/task-common/e"
	"github.com/Zoxu0928/task-common/logger"
	"github.com/Zoxu0928/task-common/tools"

	"strings"
	"time"
)

// Socket处理器，实现Socket只需要实现Socket处理器即可
type Handler struct {
	codec              Codec
	OnSessionConnected func(session *Session)
	OnSessionClosed    func(session *Session)
	OnMessageReceived  func(session *Session, message *Msg)
	OnException        func(session *Session)
}

// 处理网络输入流共用方法
func (handler *Handler) readIo(socket *socket, session *Session) {

	defer e.OnError("handler.readIo")

	flag := "server"
	if session.isClient {
		flag = "client"
	}

	conn := session.connection

	// 连续失败次数
	fail_times := 0

	// 缓存本次解码的消息，这个消息可能分多次解码才能完整
	last_msg := &Msg{}

	// 数据流缓冲
	buf := bytes.NewBuffer([]byte{})

	// 每次读入的新数据
	ioData := make([]byte, 1024*4)
	for {

		if socket.closed {
			logger.Error("%s [%s] [%d] Connection has closed. remote=%s. i will return.", socket.addr, flag, session.id, conn.RemoteAddr())
			return
		}

		// 向缓冲区读入io流
		len, err := conn.Read(ioData)

		// 错误判断
		if err != nil {

			// 判断是否远程连接被关闭
			if err.Error() == "EOF" {
				logger.Info("%s [%s] [%d] Connection close. Cause of remote connection has closed. remote=%s", socket.addr, flag, session.id, conn.RemoteAddr())
				session.Close()
				return
			}
			if strings.Contains(err.Error(), "closed by the remote host") {
				logger.Info("%s [%s] [%d] Connection close. Cause of remote connection has interrupted. remote=%s", socket.addr, flag, session.id, conn.RemoteAddr())
				session.Close()
				return
			}
			if strings.Contains(err.Error(), "use of closed network") {
				logger.Info("%s [%s] [%d] Connection close. Cause of connection has closed. remote=%s", socket.addr, flag, session.id, conn.RemoteAddr())
				session.Close()
				return
			}

			// 错误后重试3次，3次后还错误则关闭连接
			fail_times++
			logger.Error("%s [%s] [%d] Read io error. times=%s. remote=%s. Cause of %s", socket.addr, flag, session.id, tools.ToString(fail_times), conn.RemoteAddr(), err.Error())
			if socket.handler.OnException != nil {
				go func() {
					defer e.OnError("")
					socket.handler.OnException(session)
				}()
			}
			if fail_times >= 3 {
				logger.Error("%s [%s] [%d] Connection close. remote=%s. Error finished.", socket.addr, flag, session.id, conn.RemoteAddr())
				session.Close()
				return
			}
			time.Sleep(200 * time.Millisecond)
			continue
		} else {
			fail_times = 0
		}

		// 如果读到了数据
		if len > 0 {

			buf.Write(ioData[0:len])

			// 缓冲区中的数据，如果粘包了则可以解码出多条消息
			for {

				// 解码成功
				if ok := socket.handler.codec.Decode(last_msg, buf); ok {

					// 复制一份消息数据
					copyMsg := *last_msg

					// 回调
					if socket.handler.OnMessageReceived != nil {
						if socket.ordered {
							socket.handler.OnMessageReceived(session, &copyMsg)
						} else {
							go func(cmsg *Msg) {
								defer e.OnError("")
								socket.handler.OnMessageReceived(session, cmsg)
							}(&copyMsg)
						}
					}

					// 下次要重新初始化
					last_msg = &Msg{}

					// 断包了，待下次再解码
				} else {
					break
				}
			}
		}
	}
}
