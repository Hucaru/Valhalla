package game

import (
	"fmt"
	"math/rand"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type fieldRectangle struct {
	left, top, right, bottom int
}

func (r fieldRectangle) empty() bool {
	if (r.left | r.top | r.right | r.bottom) == 0 {
		return true
	}

	return false
}

func (r *fieldRectangle) inflate(x, y int) {
	r.left -= x
	r.top += y
	r.right += x
	r.bottom -= y
}

func (r fieldRectangle) width() int {
	return r.right - r.left
}

func (r fieldRectangle) height() int {
	return r.top - r.bottom
}

func (r fieldRectangle) contains(x, y int) bool {
	if r.left > x {
		return false
	}

	if r.top < y {
		return false
	}

	if r.right < x {
		return false
	}

	if r.bottom > y {
		return false
	}

	return true
}

type field struct {
	id        int32
	instances []instance
	data      nx.Map
	dispatch  chan func()
	server    *ChannelServer

	vrlimit, mbr, ombr             fieldRectangle
	mobCapacityMin, mobCapacityMax int
}

func (f *field) createInstance() int {
	id := len(f.instances)
	npcs := make([]npc, len(f.data.NPCs))

	for i, l := range f.data.NPCs {
		npcs[i] = createNpc(int32(i), l)
	}

	// add initial set of mobs

	f.instances = append(f.instances, instance{
		id:      id,
		fieldID: f.id,
		npcs:    npcs,
		server:  f.server,
	})

	// register map work function

	return id
}

func (f *field) calculateFieldLimits() {

}

func (f field) validInstance(instance int) bool {
	if len(f.instances) > instance && instance > -1 {
		return true
	}
	return false
}

func (f *field) deleteInstance(instance int) error {
	if f.validInstance(instance) {
		return f.instances[instance].delete()
	}
	return fmt.Errorf("Invalid instance")
}

func (f *field) addPlayer(conn mnet.Client, instance int) error {
	if f.validInstance(instance) {
		return f.instances[instance].addPlayer(conn)
	}

	return fmt.Errorf("Invalid instance")
}

func (f *field) removePlayer(conn mnet.Client, instance int) error {
	return nil
}

func (f field) getRandomSpawnPortal() (nx.Portal, byte, error) {
	portals := []nx.Portal{}
	inds := []int{}

	nxMap, err := nx.GetMap(f.id)

	if err != nil {
		return nx.Portal{}, 0, fmt.Errorf("Invalid map id")
	}

	for i, p := range nxMap.Portals {
		if p.Pn == "sp" {
			portals = append(portals, p)
			inds = append(inds, i)
		}
	}

	ind := rand.Intn(len(portals))
	return portals[ind], byte(inds[ind]), nil
}

func (f field) send(p mpacket.Packet, instance int) error {
	if f.validInstance(instance) {
		return f.instances[instance].send(p)
	}

	return fmt.Errorf("Invalid instance")
}

func (f field) sendExcept(p mpacket.Packet, exception mnet.Client, instance int) error {
	if f.validInstance(instance) {
		return f.instances[instance].sendExcept(p, exception)
	}

	return fmt.Errorf("Invalid instance")
}
