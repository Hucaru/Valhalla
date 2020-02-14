package field

import (
	"fmt"

	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/field/mob"
	"github.com/Hucaru/Valhalla/server/field/npc"
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

// Field data
type Field struct {
	ID        int32
	instances []Instance
	Data      nx.Map

	vrlimit, mbr, ombr             fieldRectangle
	mobCapacityMin, mobCapacityMax int

	Dispatch chan func()
}

// CreateInstance for this field
func (f *Field) CreateInstance() int {
	id := len(f.instances)
	npcs := make([]npc.Data, len(f.Data.NPCs))

	for i, l := range f.Data.NPCs {
		npcs[i] = npc.CreateFromData(int32(i), l)
	}

	portals := make([]Portal, len(f.Data.Portals))
	for i, p := range f.Data.Portals {
		portals[i] = createPortalFromData(p)
		portals[i].id = byte(i)
	}

	// add initial set of mobs
	mobs := make([]mob.Data, len(f.Data.Mobs))
	for i, v := range f.Data.Mobs {
		m, err := nx.GetMob(v.ID)

		if err != nil {
			continue
		}

		mobs[i] = mob.CreateFromData(int32(i+1), v, m, true, true)
		mobs[i].SetSummonType(-1)
	}

	f.instances = append(f.instances, Instance{
		id:       id,
		fieldID:  f.ID,
		npcs:     npcs,
		portals:  portals,
		dispatch: f.Dispatch,
		mobs:     mobs,
	})

	return id
}

// CalculateFieldLimits for mob spawning
func (f *Field) CalculateFieldLimits() {

}

func (f Field) validInstance(instance int) bool {
	if len(f.instances) > instance && instance > -1 {
		return true
	}
	return false
}

// DeleteInstance from id
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

// GetInstance from id
func (f *Field) GetInstance(id int) (*Instance, error) {
	if f.validInstance(id) {
		return &f.instances[id], nil
	}

	return nil, fmt.Errorf("Invalid instance id")
}

// Instances in field
func (f *Field) Instances() []Instance {
	return f.instances
}

// ChangePlayerInstance id
func (f *Field) ChangePlayerInstance(player player, id int) error {
	if id == player.InstanceID() {
		return fmt.Errorf("In specified instance")
	}

	if f.validInstance(id) {
		err := f.instances[player.InstanceID()].RemovePlayer(player)

		if err != nil {
			return err
		}

		player.SetInstance(f.instances[id])
		err = f.instances[id].AddPlayer(player)

		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("Invalid instance id")
}
