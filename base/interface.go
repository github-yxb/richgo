package base

type IModule interface {
	Init() bool
}

type ICallRemote interface {
	Call(moduleTag, method string, args ...interface{}) interface{}
	Send(moduleTag, method string, args ...interface{}) error
}

type IPacket interface {
	Pack(cmd uint32, data []byte) []byte
	HeadLen() uint32
	HeadInfo(headBuf []byte) (uint32, uint32)
}

type IAgent interface {
	Send(cmd uint32, data []byte)
	Close()
	Run()
}

type IGateRouter interface {
	GetModuleTagByCmd(cmd uint32) string
}


type IGateHandler interface {
	HandleData(conn IAgent, cmd uint32, data []byte)
}