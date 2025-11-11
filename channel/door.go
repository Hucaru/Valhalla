package channel

import (
	"time"

	"github.com/Hucaru/Valhalla/nx"
)

// createMysticDoor creates a mystic door for the player
func createMysticDoor(plr *Player, skillID int32, skillLevel byte) {
	if plr.doorMapID != 0 {
		removeMysticDoor(plr)
	}

	doorPos := plr.pos

	var duration int64
	if data, err := nx.GetPlayerSkill(skillID); err == nil {
		idx := int(skillLevel) - 1
		if idx >= 0 && idx < len(data) && data[idx].Time > 0 {
			duration = data[idx].Time
		}
	}

	plr.inst.send(packetPlayerSkillAnimation(plr.ID, false, skillID, skillLevel))

	expiresAt := time.Now().Add(time.Duration(duration) * time.Second)
	createSourceDoor(plr, doorPos, expiresAt)

	returnMapID := plr.inst.returnMapID
	if returnMapID > 0 {
		if returnField, ok := plr.inst.server.fields[returnMapID]; ok {
			if returnInst, err := returnField.getInstance(0); err == nil {
				createTownDoor(plr, returnInst, doorPos, expiresAt)
			}
		}
	}
}

func createSourceDoor(plr *Player, doorPos pos, expiresAt time.Time) {
	doorSpawnID := plr.inst.idCounter
	plr.inst.idCounter++

	plr.doorMapID = plr.mapID
	plr.doorSpawnID = doorSpawnID

	newPortalID := plr.inst.getNextPortalID()
	sourcePortal := portal{
		id:          newPortalID,
		pos:         doorPos,
		name:        "tp",
		destFieldID: plr.inst.returnMapID,
		destName:    "tp",
		temporary:   true,
	}
	plr.doorPortalIndex = plr.inst.addPortal(sourcePortal)

	plr.inst.mysticDoors[plr.ID] = &mysticDoorInfo{
		ownerID:     plr.ID,
		spawnID:     doorSpawnID,
		portalIndex: plr.doorPortalIndex,
		pos:         doorPos,
		destMapID:   plr.inst.returnMapID,
		townPortal:  false,
		srcPos:      doorPos,
		expiresAt:   expiresAt,
	}

	plr.inst.send(packetMapSpawnMysticDoor(doorSpawnID, doorPos, true))
	plr.inst.send(packetMapPortal(plr.mapID, plr.inst.returnMapID, doorPos))

	if plr.party != nil {
		for _, viewer := range plr.inst.players {
			if viewer == nil || viewer.party == nil || viewer.party.ID != plr.party.ID {
				continue
			}
			ownerIdx := byte(0)
			for i, pid := range viewer.party.PlayerID {
				if pid == plr.ID {
					ownerIdx = byte(i)
					break
				}
			}
			viewer.Send(packetMapPortalParty(ownerIdx, plr.mapID, plr.inst.returnMapID, doorPos))
		}
	}
}

// createTownDoor creates the door in the town map
func createTownDoor(plr *Player, townInst *fieldInstance, doorPos pos, expiresAt time.Time) {
	townPortalIdx, townPortal, err := townInst.findAvailableTownPortal()
	if err != nil {
		return
	}

	townDoorSpawnID := townInst.idCounter
	townInst.idCounter++

	plr.townDoorMapID = townInst.fieldID
	plr.townDoorSpawnID = townDoorSpawnID
	plr.townPortalIndex = townPortalIdx

	townInst.portals[townPortalIdx].destFieldID = plr.mapID
	townInst.portals[townPortalIdx].destName = "sp"
	townInst.portals[townPortalIdx].temporary = true

	townInst.mysticDoors[plr.ID] = &mysticDoorInfo{
		ownerID:     plr.ID,
		spawnID:     townDoorSpawnID,
		portalIndex: townPortalIdx,
		pos:         townPortal.pos,
		destMapID:   plr.mapID,
		townPortal:  true,
		srcPos:      doorPos,
		expiresAt:   expiresAt,
	}

	for _, viewer := range townInst.players {
		if viewer == nil {
			continue
		}
		viewer.Send(packetMapSpawnMysticDoor(townDoorSpawnID, townPortal.pos, false))
		viewer.Send(packetMapPortal(plr.mapID, townInst.fieldID, doorPos))
	}

	if plr.party != nil {
		for _, viewer := range townInst.players {
			if viewer == nil || viewer.party == nil || viewer.party.ID != plr.party.ID {
				continue
			}
			ownerIdx := byte(0)
			for i, pid := range viewer.party.PlayerID {
				if pid == plr.ID {
					ownerIdx = byte(i)
					break
				}
			}
			viewer.Send(packetMapPortalParty(ownerIdx, plr.mapID, townInst.fieldID, doorPos))
		}
	}

	plr.Send(packetMapPortal(plr.mapID, townInst.fieldID, doorPos))
	if plr.party != nil {
		ownerIdx := byte(0)
		for i, pid := range plr.party.PlayerID {
			if pid == plr.ID {
				ownerIdx = byte(i)
				break
			}
		}
		plr.Send(packetMapPortalParty(ownerIdx, plr.mapID, townInst.fieldID, doorPos))
	}
}

// removeMysticDoor removes a player's existing mystic door
func removeMysticDoor(plr *Player) {
	if plr.doorMapID != 0 {
		if doorField, ok := plr.inst.server.fields[plr.doorMapID]; ok {
			if doorInst, err := doorField.getInstance(plr.inst.id); err == nil {
				removeDoorFromInstance(doorInst, plr.doorSpawnID, plr.doorPortalIndex, plr.ID)
			}
		}
	}

	if plr.townDoorMapID != 0 {
		if townField, ok := plr.inst.server.fields[plr.townDoorMapID]; ok {
			if townInst, err := townField.getInstance(0); err == nil {
				removeDoorFromInstance(townInst, plr.townDoorSpawnID, -1, plr.ID)
			}
		}
	}

	// Reset player-side door state
	plr.doorMapID = 0
	plr.doorSpawnID = 0
	plr.doorPortalIndex = 0
	plr.townDoorMapID = 0
	plr.townDoorSpawnID = 0
	plr.townPortalIndex = 0
}

// removeDoorFromInstance removes the door object and optional portal from an instance and broadcasts removal
func removeDoorFromInstance(inst *fieldInstance, spawnID int32, portalIndex int, ownerID int32) {
	inst.send(packetMapRemoveMysticDoor(spawnID, true))

	if portalIndex >= 0 {
		inst.removePortalAtIndex(portalIndex)
	}

	removePkt := packetMapRemovePortal()
	inst.send(removePkt)

	delete(inst.mysticDoors, ownerID)

	for _, p := range inst.players {
		if p != nil && p.ID == ownerID {
			p.doorMapID = 0
			p.doorSpawnID = 0
			p.doorPortalIndex = 0
			p.townDoorMapID = 0
			p.townDoorSpawnID = 0
			p.townPortalIndex = 0
			break
		}
	}
}

// removeMysticDoorByIDs removes a player's door by IDs
func removeMysticDoorByIDs(server *Server, ownerID, sourceMapID, townMapID int32) {
	if sourceField, ok := server.fields[sourceMapID]; ok {
		if sourceInst, err := sourceField.getInstance(0); err == nil {
			if doorInfo, exists := sourceInst.mysticDoors[ownerID]; exists {
				removeDoorFromInstance(sourceInst, doorInfo.spawnID, doorInfo.portalIndex, ownerID)
			}
		}
	}

	if townMapID > 0 {
		if townField, ok := server.fields[townMapID]; ok {
			if townInst, err := townField.getInstance(0); err == nil {
				if doorInfo, exists := townInst.mysticDoors[ownerID]; exists {
					removeDoorFromInstance(townInst, doorInfo.spawnID, -1, ownerID)
				}
			}
		}
	}
}
