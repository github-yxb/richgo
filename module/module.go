package module

import (
	"time"
)
import "richGo/actor"
import "richGo/base"

type ModuleFactory func(base.ICallRemote, map[string]interface{}) base.IModule

type ActorModule struct {
	actor.Actor
	Tag string
	Name string
	Node base.ICallRemote
}

func (m *ActorModule) After(d time.Duration, t func()) *time.Timer {
	return time.AfterFunc(d, func() {
		m.Post(t)
	})
}



