package entity

import (
	"fmt"

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

type Field struct {
	ID        int32
	instances []instance
	Data      nx.Map
	Players   *Players

	vrlimit, mbr, ombr             fieldRectangle
	mobCapacityMin, mobCapacityMax int
}

func (f *Field) CreateInstance() int {
	id := len(f.instances)
	npcs := make([]npc, len(f.Data.NPCs))

	for i, l := range f.Data.NPCs {
		npcs[i] = createNpcFromData(int32(i), l)
	}

	portals := make([]portal, len(f.Data.Portals))
	for i, p := range f.Data.Portals {
		portals[i] = createPortalFromData(p)
		portals[i].id = byte(i)
	}

	// add initial set of mobs

	f.instances = append(f.instances, instance{
		id:      id,
		fieldID: f.ID,
		npcs:    npcs,
		players: f.Players,
		portals: portals,
	})

	// register map work function

	return id
}

func (f *Field) CalculateFieldLimits() {

}

func (f Field) validInstance(instance int) bool {
	if len(f.instances) > instance && instance > -1 {
		return true
	}
	return false
}

func (f *Field) DeleteInstance(instance int) error {
	if f.validInstance(instance) {
		return f.instances[instance].delete()
	}
	return fmt.Errorf("Invalid instance")
}

func (f *Field) AddPlayer(conn mnet.Client, instance int) error {
	if f.validInstance(instance) {
		return f.instances[instance].addPlayer(conn)
	}

	return fmt.Errorf("Invalid instance")
}

func (f *Field) RemovePlayer(conn mnet.Client, instance int) error {
	if f.validInstance(instance) {
		return f.instances[instance].removePlayer(conn)
	}

	return fmt.Errorf("Invalid instance")
}

func (f Field) GetRandomSpawnPortal() (portal, error) { // spawn portals are instance independent
	if len(f.instances) < 1 {
		return portal{}, fmt.Errorf("No instances in map")
	}

	return f.instances[0].getRandomSpawnPortal()
}

func (f Field) GetPortalFromName(name string) (portal, error) { // only spawn portals have string names
	if len(f.instances) < 1 {
		return portal{}, fmt.Errorf("No instances in map")
	}

	return f.instances[0].getPortalFromName(name)
}

func (f Field) Send(p mpacket.Packet, instance int) error {
	if f.validInstance(instance) {
		return f.instances[instance].send(p)
	}

	return fmt.Errorf("Invalid instance")
}

func (f Field) SendExcept(p mpacket.Packet, exception mnet.Client, instance int) error {
	if f.validInstance(instance) {
		return f.instances[instance].sendExcept(p, exception)
	}

	return fmt.Errorf("Invalid instance")
}
