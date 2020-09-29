package richGo

import (
	"context"
	"fmt"
	"github.com/github-yxb/richgo/module"
	"reflect"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	rClient "github.com/smallnest/rpcx/client"
	rServer "github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"github.com/spf13/viper"
)

type RemoteArgs struct {
	moduleTag    string
	moduleMethod string
	needReply    bool
	args         []interface{}
}

type RemoteReply struct {
	reply []interface{}
}

const ETCD_BASE_PATH = "/richgo_modules"

var once sync.Once

//Node 在进程中只有一个，管理modules,存储配置等
type Node struct {
	config        *viper.Viper
	moduleManager *module.ModuleManager
	rpcServer     *rServer.Server
	rpcClient     *rClient.OneClient
}

var node *Node

func GetNode(config *viper.Viper) *Node {
	once.Do(func() {
		node = &Node{config: config, moduleManager: module.NewModuleManager()}
	})

	return node
}

func (n *Node) AddModule(module string, factory module.ModuleFactory) {
	n.moduleManager.AddModeFactory(module, factory)
}

// 注册所有module到etcd
func (n *Node) RegisterModules() error {

	for _, tag := range n.moduleManager.GetAllModuleTags() {
		if err := n.rpcServer.RegisterFunction(tag, n.RemoteCall, ""); err != nil {
			return err
		}
	}

	return nil

}

func (n *Node) InitRClient() {
	etcdAddr := n.config.GetStringSlice("etcd")
	d := rClient.NewEtcdDiscoveryTemplate(ETCD_BASE_PATH, etcdAddr, nil)
	n.rpcClient = rClient.NewOneClient(rClient.Failfast, rClient.SelectByUser, d, rClient.DefaultOption)
}

func (n *Node) Start(nodeName string) {
	config := viper.GetViper()
	if err := n.moduleManager.InitModule(n, nodeName, config); err != nil {
		logrus.Errorf("init module failed:%s", err.Error())
		return
	}
	etcdAddr := config.GetStringSlice("etcd")
	rpcPort := config.GetString(fmt.Sprintf("nodes.%s.rpc_port", nodeName))
	rpcAddress := "0.0.0.0:" + rpcPort

	n.rpcServer = rServer.NewServer()
	etcd := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: "tcp@" + rpcAddress,
		EtcdServers:    etcdAddr,
		BasePath:       ETCD_BASE_PATH,
		UpdateInterval: time.Second * 5,
	}

	if err := etcd.Start(); err != nil {
		logrus.Errorf("etcd start failed with:%s, error:%s", etcdAddr, err.Error())
		return
	}
	n.rpcServer.Plugins.Add(etcd)

	if err := n.RegisterModules(); err != nil {
		n.rpcServer.UnregisterAll()
		logrus.Errorf("register modules error:%s", err.Error())
		return
	}

	n.InitRClient()

	if err := n.rpcServer.Serve("tcp", rpcAddress); err != nil {
		logrus.Errorf("rpc server server failed with:%s, error:%s", rpcAddress, err.Error())
		n.rpcServer.UnregisterAll()
		return
	}

}

func (n *Node) RemoteCall(ctx context.Context, args *RemoteArgs, reply *RemoteReply) error {
	if args == nil || reply == nil {
		return fmt.Errorf("params error")
	}

	targetModule := n.moduleManager.GetModuleByTag(args.moduleTag)
	if targetModule == nil {
		return fmt.Errorf("can't find module by tag:%s", args.moduleTag)
	}

	moduleType := reflect.TypeOf(targetModule)
	method, find := moduleType.MethodByName(args.moduleMethod)
	if !find {
		return fmt.Errorf("can't find moduld:%s method:%s", args.moduleTag, args.moduleMethod)
	}

	if method.Type.NumIn()-1 != len(args.args) {
		return fmt.Errorf("number of method params error, need-%d, args-%d", method.Type.NumIn()-1, len(args.args))
	}

	var moduleCallFuncName string
	var inFunc interface{}

	tmpInFunc := func() interface{} {
		inArgs := make([]reflect.Value, 0)
		inArgs = append(inArgs, reflect.ValueOf(targetModule))
		for _, item := range args.args {
			inArgs = append(inArgs, reflect.ValueOf(item))
		}

		return method.Func.Call(inArgs)
	}

	if args.needReply {
		moduleCallFuncName = "Call"
		inFunc = tmpInFunc
	} else {
		moduleCallFuncName = "Post"
		inFunc = func() {
			tmpInFunc()
		}
	}
	moduleCallFunc, find := moduleType.MethodByName(moduleCallFuncName)
	if !find {
		return fmt.Errorf("invalid module, module not have method-%s", moduleCallFuncName)
	}

	inArgs := []reflect.Value{reflect.ValueOf(targetModule), reflect.ValueOf(inFunc)}
	ret := moduleCallFunc.Func.Call(inArgs)

	if len(ret) != 0 {
		retSlice := ret[0].Interface().([]reflect.Value)
		// logrus.Info("real ret:", retSlice)
		if len(retSlice) != 0 {
			err, ok := retSlice[0].Interface().(error)
			// logrus.Info("error convert:", err, ok)
			if ok {
				return err
			} else {
				for i := 0; i < len(retSlice); i++ {
					reply.reply = append(reply.reply, retSlice[i].Interface())
				}
			}
		}

	}

	return nil
}

// 调用远程模块的方法,需要返回值
func (n *Node) Call(moduleTag, method string, args ...interface{}) interface{} {

	remoteArg := RemoteArgs{
		moduleTag:    moduleTag,
		moduleMethod: method,
		args:         args,
		needReply:    true,
	}

	reply := RemoteReply{}

	if err := n.rpcClient.Call(context.Background(), moduleTag, "RemoteCall", remoteArg, reply); err != nil {
		return err
	}

	return reply.reply
}

// 调用远程模块的方法，无需返回值
func (n *Node) Send(moduleTag, method string, args ...interface{}) error {

	remoteArg := RemoteArgs{
		moduleTag:    moduleTag,
		moduleMethod: method,
		args:         args,
		needReply:    true,
	}

	reply := RemoteReply{}
	if _, err := n.rpcClient.Go(context.Background(), moduleTag, "RemoteCall", remoteArg, reply, nil); err != nil {
		return err
	}

	return nil
}
