package script

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

type scriptStore struct {
	scripts map[string]scriptFile
	mutex   *sync.RWMutex
}

var loadedScripts = scriptStore{
	scripts: make(map[string]scriptFile),
	mutex:   &sync.RWMutex{},
}

func Get(name string) (string, error) {
	return loadedScripts.get(name)
}

func (ss *scriptStore) add(s scriptFile) {
	ss.mutex.Lock()
	ss.scripts[strings.TrimSuffix(s.name, filepath.Ext(s.name))] = s
	ss.mutex.Unlock()
}

func (ss *scriptStore) remove(name string) {
	ss.mutex.Lock()
	delete(ss.scripts, name)
	ss.mutex.Unlock()
}

func (ss *scriptStore) get(name string) (string, error) {
	s := ""
	var err error
	ss.mutex.RLock()
	if v, ok := ss.scripts[name]; ok {
		s = v.contents
	} else {
		err = fmt.Errorf("Error in getting script with name " + name)
	}
	ss.mutex.RUnlock()

	return s, err
}
