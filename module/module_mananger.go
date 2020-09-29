package module

import (
	"fmt"
	"github.com/github-yxb/richgo/base"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"sync"
)

type ModuleManager struct {
	moduleFactoryMap map[string]ModuleFactory
	modules          map[string]base.IModule
	rwLock           sync.RWMutex
}

func NewModuleManager() *ModuleManager {
	return &ModuleManager{
		moduleFactoryMap: make(map[string]ModuleFactory),
		modules:          make(map[string]base.IModule),
	}
}

func (m *ModuleManager) AddModeFactory(module string, f ModuleFactory) {
	m.moduleFactoryMap[module] = f
}

func (m *ModuleManager) GetModuleByTag(moduleTag string) base.IModule {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()

	if im, ok := m.modules[moduleTag]; ok {
		return im
	}

	return nil
}

func (m *ModuleManager) GetAllModuleTags() []string {

	var tags []string
	for tag, _ := range m.modules {
		tags = append(tags, tag)
	}

	return tags
}

func (m *ModuleManager) InitModule(n base.ICallRemote, nodeName string, v *viper.Viper) error {

	if v == nil || nodeName == "" {
		return fmt.Errorf("can't read config")
	}

	configKey := fmt.Sprintf("nodes.%s.modules", nodeName)
	modules := v.GetStringMap(configKey)
	if modules == nil || len(modules) == 0 {
		return fmt.Errorf("can't read path %s", configKey)
	}

	for moduleTag, moduleConfig := range modules {
		mapConfig := moduleConfig.(map[string]interface{})
		moduleName := mapConfig["module"].(string)

		if f, ok := m.moduleFactoryMap[moduleName]; ok {

			mapConfig["tag"] = moduleTag
			m.modules[moduleTag] = f(n, mapConfig)
			if !m.modules[moduleTag].Init() {
				return fmt.Errorf("init module failed, %s", moduleName)
			}
			logrus.Infof("module:%s initialized", moduleTag)

		} else {
			return fmt.Errorf("can't find module:%s factory", moduleName)
		}
	}

	return nil
}
