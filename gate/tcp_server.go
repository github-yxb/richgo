package gate

import (
	"github.com/github-yxb/richgo/base"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

type TcpServer struct {
	connMap  map[net.Conn]bool
	connLock sync.Mutex
	packet   base.IPacket
	handler  base.IGateHandler
}

func NewTcpServer(packet base.IPacket, handler base.IGateHandler) *TcpServer {
	s := &TcpServer{
		connMap:  make(map[net.Conn]bool),
		packet:   packet,
		handler:  handler,
		connLock: sync.Mutex{},
	}

	return s
}

func (s *TcpServer) Start(addr string) error {

	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		logrus.Errorf("listen addr:%s failed:%s", addr, err.Error())
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			logrus.Errorf("accept error:%s", err.Error())
			continue
		}

		logrus.Debugf("new connection:%s", conn.RemoteAddr().String())

		s.connLock.Lock()
		s.connMap[conn] = true
		s.connLock.Unlock()

		go func() {
			a := NewAgent(conn)
			a.Run()

			s.connLock.Lock()
			delete(s.connMap, a.conn)
			s.connLock.Unlock()
		}()
	}
}
