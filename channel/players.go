package channel

import (
	"log"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/interop"
)

var Players = maplePlayers{players: make(map[interop.ClientConn]*MapleCharacter), mutex: &sync.RWMutex{}}

func init() {
	go func() {
		// Save character data every 15 mins
		ticker := time.NewTicker(15 * time.Minute)

		for {
			<-ticker.C

			Players.OnCharacters(func(char *MapleCharacter) {
				err := char.Save()

				if err != nil {
					log.Println("Unable to save character data")
				}
			})

		}
	}()
}

type maplePlayers struct {
	players map[interop.ClientConn]*MapleCharacter // keep as maps as it's sparse data?
	mutex   *sync.RWMutex
}

func (p *maplePlayers) AddPlayer(conn interop.ClientConn, char *character.Character) {
	p.mutex.Lock()
	p.players[conn] = &MapleCharacter{*char, conn}
	p.mutex.Unlock()
}

func (p *maplePlayers) RemovePlayer(conn interop.ClientConn) {
	p.mutex.Lock()
	if _, exists := p.players[conn]; exists {
		delete(p.players, conn)
	}
	p.mutex.Unlock()
}

func (p *maplePlayers) OnCharacters(action func(char *MapleCharacter)) {
	p.mutex.RLock()
	for _, value := range p.players {
		action(value)
	}
	p.mutex.RUnlock()
}

func (p *maplePlayers) OnCharacterFromConn(conn interop.ClientConn, action func(char *MapleCharacter)) {
	p.mutex.RLock()
	if _, exists := p.players[conn]; exists {
		action(p.players[conn])
	}
	p.mutex.RUnlock()
}

func (p *maplePlayers) OnCharacterFromName(name string, action func(char *MapleCharacter)) {
	p.mutex.RLock()
	for _, char := range p.players {
		if char.GetName() == name {
			action(char)
			break
		}
	}
	p.mutex.RUnlock()
}

func (p *maplePlayers) OnCharacterFromID(id int32, action func(char *MapleCharacter)) {
	p.mutex.RLock()
	for _, char := range p.players {
		if char.GetCharID() == id {
			action(char)
			break
		}
	}
	p.mutex.RUnlock()
}
