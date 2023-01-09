package manager

import (
	"sync"
)

type GoroutineManager struct {
	m sync.Map
}

func Init() *GoroutineManager {
	return new(GoroutineManager)
}

func (g *GoroutineManager) Add(ch chan bool, key string) {
	g.m.Store(key, ch)
}

func (g *GoroutineManager) Remove(key string) {
	q, exist := g.m.Load(key)
	if exist {
		q.(chan bool) <- true
		//close(q.(chan bool))
		g.m.Delete(key)
	}
}

func (g *GoroutineManager) Get(key string) bool {
	_, exist := g.m.Load(key)
	return exist
}

func (g *GoroutineManager) ClearAll() {

	g.m.Range(func(key interface{}, value interface{}) bool {
		q, exist := g.m.Load(key)
		if exist {
			q.(chan bool) <- true
		}
		g.m.Delete(key)
		return true
	})
}
