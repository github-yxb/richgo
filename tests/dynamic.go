package main

import (
	"fmt"
	"reflect"
	"github.com/vmihailenco/msgpack/v5"
)

type IModule interface {
	Init()
}

type M struct {
}

func (m *M) Init() {
	fmt.Println("called")
}

type Args struct {
	V []interface{}
}

func main() {
	var im IModule = &M{}

	fmt.Println(reflect.TypeOf(im), reflect.ValueOf(im))
	objValue := reflect.ValueOf(im)
	objType := reflect.TypeOf(im)
	method, find := objType.MethodByName("Init")
	if find {
		method.Func.Call([]reflect.Value{objValue})
	}

	msgpack.Register(im, nil, nil)
	arg := Args{V : []interface{}{1, 3.123, "string", false, map[string]int{"a" : 2}}}
	b, err := msgpack.Marshal(arg)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(b, arg.V)

	arg2 := Args{}
	msgpack.Unmarshal(b, &arg2)
	fmt.Println(arg2.V)

}
