package channel

import (
	"log"
	"sync"

	"github.com/Hucaru/Valhalla/mnet"
)

type sessionRegistry struct {
	mu   sync.RWMutex
	mapC map[mnet.Client]*player
}

var sessions = &sessionRegistry{mapC: make(map[mnet.Client]*player)}

func attachSession(conn mnet.Client, p *player) {
	if conn == nil || p == nil {
		return
	}
	sessions.mu.Lock()
	sessions.mapC[conn] = p
	sessions.mu.Unlock()
}

func detachSession(conn mnet.Client) *player {
	if conn == nil {
		return nil
	}
	sessions.mu.Lock()
	p := sessions.mapC[conn]
	delete(sessions.mapC, conn)
	sessions.mu.Unlock()
	return p
}

// OnDisconnect must be called when the socket closes.
// It flushes coalesced updates and performs a full checkpoint save.
func OnDisconnect(conn mnet.Client) {
	if conn == nil {
		return
	}
	plr := detachSession(conn)
	if plr == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic during OnDisconnect logout for player %d: %v", plr.id, r)
		}
	}()
	plr.Logout("disconnect")
}
