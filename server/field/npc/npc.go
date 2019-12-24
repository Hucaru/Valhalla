package npc

import (
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/pos"
)

// Controller of the npc
type Controller interface {
	Send(mpacket.Packet)
}

// Data for an npc in a field
type Data struct {
	controller Controller
	id         int32
	spawnID    int32
	pos        pos.Data
	foothold   int16
	faceLeft   bool
	rx0, rx1   int16
}

// CreateFromData - creates npc from nx data
func CreateFromData(spawnID int32, life nx.Life) Data {
	return Data{id: life.ID,
		spawnID:  spawnID,
		pos:      pos.New(life.X, life.Y),
		foothold: life.Foothold,
		faceLeft: life.FaceLeft,
		rx0:      life.Rx0,
		rx1:      life.Rx1}
}

// Controller of npc
func (d Data) Controller() Controller {
	return d.controller
}

// ID of npc
func (d Data) ID() int32 {
	return d.id
}

// SpawnID of npc
func (d Data) SpawnID() int32 {
	return d.spawnID
}

// Pos of npc
func (d Data) Pos() pos.Data {
	return d.pos
}

// Foothold the npc is attached to
func (d Data) Foothold() int16 {
	return d.foothold
}

// FaceLeft - does npc face left direction
func (d Data) FaceLeft() bool {
	return d.faceLeft
}

// Rx0 of npc
func (d Data) Rx0() int16 {
	return d.rx0
}

// Rx1 of npc
func (d Data) Rx1() int16 {
	return d.rx1
}

// SetController of npc
func (d *Data) SetController(controller Controller) {
	d.controller = controller
	controller.Send(packetNpcSetController(d.spawnID, true))
}

// RemoveController from npc
func (d *Data) RemoveController() {
	if d.controller != nil {
		d.controller.Send(packetNpcSetController(d.spawnID, false))
	}

	d.controller = nil
}

type instance interface {
	Send(mpacket.Packet) error
}

// AcknowledgeController movement data
func (d Data) AcknowledgeController(plr Controller, inst instance, data []byte) {
	if d.controller != plr {
		plr.Send(packetNpcSetController(d.spawnID, false))
	}

	inst.Send(packetNpcMovement(data))
}
