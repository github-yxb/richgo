package gate

import (
	"bytes"
	"encoding/binary"
	"github.com/github-yxb/richgo/base"
	"github.com/github-yxb/richgo/module"
	"github.com/sirupsen/logrus"
)

type DefaultPacket struct {
}

func (p *DefaultPacket) Pack(cmd uint32, data []byte) []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, cmd)
	binary.Write(buf, binary.LittleEndian, p.HeadLen()+uint32(len(data)))
	binary.Write(buf, binary.LittleEndian, data)

	return buf.Bytes()
}

func (p *DefaultPacket) HeadLen() uint32 {
	return 4
}

func (p *DefaultPacket) HeadInfo(headBuf []byte) (uint32, uint32) {
	var cmd, length uint32

	r := bytes.NewReader(headBuf)
	binary.Read(r, binary.LittleEndian, &cmd)
	binary.Read(r, binary.LittleEndian, &length)

	return cmd, length - p.HeadLen()
}

type Gate struct {
	module.ActorModule
	tcpServer  *TcpServer
	listenAddr string
	gateConfig map[string]interface{}
	router     base.IGateRouter
}

func NewGateModule(config map[string]interface{}) base.IModule {

	g := &Gate{
		gateConfig: config,
	}
	g.tcpServer = NewTcpServer(&DefaultPacket{}, g)
	return g
}

func (g *Gate) Init() bool {
	return true
}

func (g *Gate) Start() {
	listenAddr, f := g.gateConfig["listen_addr"]
	if !f {
		logrus.Error("gate config listen_addr not find")
		return
	}

	g.tcpServer.Start(listenAddr.(string))
}

func (g *Gate) HandleData(agent base.IAgent, cmd uint32, data []byte) {

}
