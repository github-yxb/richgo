package gate

import (
	"github.com/github-yxb/richgo/base"
	"github.com/sirupsen/logrus"
	"io"
	"net"
)

type Agent struct {
	conn    net.Conn
	sendCh  chan []byte
	packet  base.IPacket
	handler base.IGateHandler
}

func NewAgent(conn net.Conn) *Agent {
	a := &Agent{
		conn:   conn,
		sendCh: make(chan []byte),
	}

	return a
}

func (a *Agent) Run() {

	go a.sendTask()

	headBuf := make([]byte, a.packet.HeadLen())
	for {

		if _, err := io.ReadFull(a.conn, headBuf); err != nil {
			logrus.Debugf("read error:%s", err.Error())
			break
		}

		cmd, dataLen := a.packet.HeadInfo(headBuf)
		logicBuf := make([]byte, dataLen)
		if _, err := io.ReadFull(a.conn, logicBuf); err != nil {
			logrus.Debugf("read error:%s", err.Error())
			break
		}

		a.handler.HandleData(a, cmd, logicBuf)
	}
}

func (a *Agent) Send(cmd uint32, data []byte) {
	packed := a.packet.Pack(cmd, data)
	a.sendCh <- packed
}

func (a *Agent) sendTask() {
	for data := range a.sendCh {
		if _, err := a.conn.Write(data); err != nil {
			logrus.Debugf("write error:%s", err.Error())
			break
		}
	}
}

func (a *Agent) Close() {
	a.conn.Close()
	close(a.sendCh)
}
