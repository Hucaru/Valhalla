package field

import (
	"fmt"

	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/field/lifepool"
)

// Field data
type Field struct {
	ID        int32
	instances []*Instance
	Data      nx.Map

	deltaX, deltaY float64

	Dispatch chan func()
}

// CreateInstance for this field
func (f *Field) CreateInstance() int {
	id := len(f.instances)

	portals := make([]Portal, len(f.Data.Portals))
	for i, p := range f.Data.Portals {
		portals[i] = createPortalFromData(p)
		portals[i].id = byte(i)
	}

	inst := &Instance{
		id:          id,
		fieldID:     f.ID,
		portals:     portals,
		dispatch:    f.Dispatch,
		town:        f.Data.Town,
		returnMapID: f.Data.ReturnMap,
		timeLimit:   f.Data.TimeLimit,
	}

	lifePool := lifepool.CreatNewPool(inst, f.Data.NPCs, f.Data.Mobs, f.deltaX, f.deltaY, f.Data.MobRate)

	inst.lifePool = lifePool

	f.instances = append(f.instances, inst)

	return id
}

// CalculateFieldLimits for mob spawning
func (f *Field) CalculateFieldLimits() {
	// not sure if this is correct
	f.deltaX = float64(f.Data.VRRight - f.Data.VRLeft)
	f.deltaY = float64(f.Data.VRTop - f.Data.VRBottom)
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
		if len(f.instances[id].players) > 0 {
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
		return f.instances[id], nil
	}

	return nil, fmt.Errorf("Invalid instance id")
}

// Instances in field
func (f *Field) Instances() []*Instance {
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
