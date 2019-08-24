package entity

import (
	"fmt"

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

	portals := make([]Portal, len(f.Data.Portals))
	for i, p := range f.Data.Portals {
		portals[i] = createPortalFromData(p)
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

func (f *Field) DeleteInstance(id int) error {
	if f.validInstance(id) {
		if f.instances[id].PlayerCount() > 0 {
			return fmt.Errorf("Cannot delete an instance with players in it")
		}
		err := f.instances[id].delete()

		if err != nil {
			return err
		}

		f.instances = append(f.instances[:id], f.instances[id+1:]...)

		return nil
	}
	return fmt.Errorf("Invalid instance")
}

func (f *Field) GetInstance(id int) (*instance, error) {
	if f.validInstance(id) {
		return &f.instances[id], nil
	}

	return nil, fmt.Errorf("Invalid instance id")
}

func (f *Field) Instances() []instance {
	return f.instances
}

func (f *Field) ChangePlayerInstance(player *Player, id int) error {
	if id == player.instanceID {
		return fmt.Errorf("In specified instance")
	}

	if f.validInstance(id) {
		err := f.instances[player.InstanceID()].RemovePlayer(player)

		if err != nil {
			return err
		}

		player.instanceID = id
		err = f.instances[id].AddPlayer(player)

		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("Invalid instance id")
}
