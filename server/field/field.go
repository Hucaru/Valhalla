package field

import (
	"fmt"
	"math"

	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/field/droppool"
	"github.com/Hucaru/Valhalla/server/field/lifepool"
	"github.com/Hucaru/Valhalla/server/field/rectangle"
	"github.com/Hucaru/Valhalla/server/field/roompool"
)

// Field data
type Field struct {
	ID        int32
	instances []*Instance
	Data      nx.Map

	deltaX, deltaY float64

	Dispatch chan func()

	vrLimit                        rectangle.Data
	mobCapacityMin, mobCapacityMax int
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
		properties:  make(map[string]interface{}),
	}

	inst.roomPool = roompool.CreateNewPool(inst)
	inst.dropPool = droppool.CreateNewPool(inst)
	inst.lifePool = lifepool.CreatNewPool(inst, f.Data.NPCs, f.Data.Mobs, f.mobCapacityMin, f.mobCapacityMax)
	inst.lifePool.SetDropPool(&inst.dropPool)

	f.instances = append(f.instances, inst)

	return id
}

// CalculateFieldLimits for mob spawning
func (f *Field) CalculateFieldLimits() {
	vrLimit := rectangle.CreateFromLTRB(f.Data.VRLeft, f.Data.VRTop, f.Data.VRRight, f.Data.VRBottom)

	var left int64 = math.MaxInt32
	var top int64 = math.MaxInt32
	var right int64 = math.MinInt32
	var bottom int64 = math.MinInt32

	for _, fh := range f.Data.Footholds {
		if int64(fh.X1) < left {
			left = int64(fh.X1)
		}

		if int64(fh.Y1) < top {
			top = int64(fh.Y1)
		}

		if int64(fh.X2) < left {
			left = int64(fh.X2)
		}

		if int64(fh.Y2) < top {
			top = int64(fh.Y2)
		}

		if int64(fh.X1) > right {
			right = int64(fh.X1)
		}

		if int64(fh.Y1) > bottom {
			bottom = int64(fh.Y1)
		}

		if int64(fh.X2) > right {
			right = int64(fh.X2)
		}

		if int64(fh.Y2) > bottom {
			bottom = int64(fh.Y2)
		}
	}

	if !vrLimit.Empty() {
		f.vrLimit = vrLimit
	} else {
		f.vrLimit = rectangle.CreateFromLTRB(left, top-300, right, bottom+75)
	}

	left += 30
	top -= 300
	right -= 30
	bottom += 10

	if !vrLimit.Empty() {
		if vrLimit.Left+20 < left {
			left = vrLimit.Left + 20
		}

		if vrLimit.Top+65 < top {
			top = vrLimit.Top + 20
		}

		if vrLimit.Right-5 > right {
			right = vrLimit.Right - 5
		}

		if vrLimit.Bottom > bottom {
			bottom = vrLimit.Bottom
		}
	}

	mbr := rectangle.CreateFromLTRB(left+10, top-375, right-10, bottom+60)
	mbr = mbr.Inflate(10, 10)

	// outofBounds := mbr.Inflate(60, 60)

	var mobX, mobY int64

	if mbr.Width() > 800 {
		mobX = mbr.Width()
	} else {
		mobX = 800
	}

	if mbr.Height()-450 > 600 {
		mobY = mbr.Height() - 450
	} else {
		mobY = 600
	}

	var mobCapacityMin int = int(float64(mobX*mobY) * f.Data.MobRate * 0.0000078125)

	if mobCapacityMin < 1 {
		mobCapacityMin = 1
	} else if mobCapacityMin > 40 {
		mobCapacityMin = 40
	}

	mobCapacityMax := mobCapacityMin * 2

	f.mobCapacityMin = mobCapacityMin
	f.mobCapacityMax = mobCapacityMax
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
