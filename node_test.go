package richGo

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"richGo/base"
	"richGo/module"
	"testing"
	"time"
	"github.com/sirupsen/logrus"
)

type TestModule struct {
	module.ActorModule
}

func (m *TestModule) Init() bool {
	go func() {
		m.Start()
	}()

	return true
}

func (m *TestModule) HandleOneErr(x, y int, n string) error {
	fmt.Println("handle remote call:", x, y, n)

	return fmt.Errorf("failed handle HandleOneErr")
}

func (m *TestModule) HandleTwoRet(x, y int, n string) (error, string) {
	fmt.Println("handle remote call:", x, y, n)

	if x * y == 100 {
		return nil, n + " success"
	}

	return fmt.Errorf("failed handle two ret"), ""
}

func (m *TestModule) HandleOneRet(x, y int, n string) string {
	fmt.Println("handle remote call:", x, y, n)
	return n + " success"
}

func (m *TestModule) HandleNoRet(x, y int, n string)  {
	fmt.Println("handle remote call:", x, y, n)
}

func addModuleToNode() {

	viper.SetConfigFile("./config_example.yaml")
	viper.ReadInConfig()
	node := GetNode(viper.GetViper())

	moduleFactory := func(map[string]interface{}) base.IModule {
		return &TestModule{
			ActorModule : module.ActorModule{
				Tag: "tm1",
				Name: "TestModule",
				Node: node,
			},
		}
	}

	node.AddModule("TestModule", moduleFactory)
}

func TestAddModule(t *testing.T) {
	addModuleToNode()
}

func TestRemoteCall(t *testing.T) {
	logrus.SetOutput(os.Stdout)
	logrus.SetReportCaller(true)
	addModuleToNode()
	node := GetNode(viper.GetViper())

	go func() {
		node.Start("test_node")
	}()

	time.Sleep(time.Second * 3)  // 等待模块注册

	args := RemoteArgs{
		moduleTag: "tm1",
		moduleMethod: "HandleTwoRet",
		needReply: true,
		args: []interface{}{10, 10, "from remote"},
	}

	reply := RemoteReply{}

	err := node.RemoteCall(context.Background(), &args, &reply)
	t.Log(err, reply.reply)
	if err != nil {
		t.Fatalf("call error:%s", err.Error())
	}

	args2 := RemoteArgs{
		moduleTag: "tm1",
		moduleMethod: "HandleOneErr",
		needReply: true,
		args: []interface{}{10, 10, "from remote"},
	}

	reply2 := RemoteReply{}

	err2 := node.RemoteCall(context.Background(), &args2, &reply2)
	t.Log(err2, reply2.reply)

	args3 := RemoteArgs{
		moduleTag: "tm1",
		moduleMethod: "HandleOneRet",
		needReply: true,
		args: []interface{}{10, 10, "from remote"},
	}

	reply3 := RemoteReply{}

	err3 := node.RemoteCall(context.Background(), &args3, &reply3)
	t.Log(err3, reply3.reply)
}