package channel

import (
	"time"

	"github.com/Hucaru/Valhalla/constant"
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

	returnMapID := plr.inst.returnMapID
	if returnMapID != constant.InvalidMap {
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, false, skillID, skillLevel))

		expiresAt := time.Now().Add(time.Duration(duration) * time.Second)
		createSourceDoor(plr, doorPos, expiresAt)

		if returnField, ok := plr.inst.server.fields[returnMapID]; ok {
			if returnInst, err := returnField.getInstance(0); err == nil {
				createTownDoor(plr, returnInst, doorPos, expiresAt)
			}
		}
	}
}

func createSourceDoor(plr *Player, doorPos pos, expiresAt time.Time) {
	doorSpawnID := plr.inst.nextID()

	plr.doorMapID = plr.mapID
	plr.doorSpawnID = doorSpawnID

	plr.doorPortalIndex = plr.inst.createNewPortal(doorPos, "tp", plr.inst.returnMapID, plr.Name, true)

	plr.inst.mysticDoors[plr.ID] = newMysticDoor(plr.ID, doorSpawnID, plr.doorPortalIndex, doorPos, doorPos, plr.inst.returnMapID, false, expiresAt)

	plr.inst.send(packetMapSpawnMysticDoor(doorSpawnID, doorPos, true))
	plr.inst.send(packetMapPortal(plr.mapID, plr.inst.returnMapID, doorPos))

	if plr.party != nil {
		for index, viewer := range plr.party.players {
			if viewer == nil {
				continue
			}

			viewer.Send(packetMapPortalParty(byte(index), plr.mapID, plr.inst.returnMapID, doorPos))
		}
	}
}

// createTownDoor creates the door in the town map
func createTownDoor(plr *Player, townInst *fieldInstance, doorPos pos, expiresAt time.Time) {
	if existing, ok := townInst.mysticDoors[plr.ID]; ok && existing != nil {
		if existing.portalIndex >= 0 && existing.portalIndex < len(townInst.portals) {
			townInst.portals[existing.portalIndex].resetTownPortal()
		}
		removeDoorFromInstance(townInst, existing.spawnID, -1, plr.ID)
	}

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
	townInst.portals[townPortalIdx].destName = "tp"
	townInst.portals[townPortalIdx].name = plr.Name
	townInst.portals[townPortalIdx].temporary = true

	townInst.mysticDoors[plr.ID] = newMysticDoor(plr.ID, townDoorSpawnID, townPortalIdx, townPortal.pos, doorPos, plr.mapID, true, expiresAt)

	for _, viewer := range townInst.players {
		if viewer == nil {
			continue
		}
		viewer.Send(packetMapSpawnMysticDoor(townDoorSpawnID, townPortal.pos, false))
		viewer.Send(packetMapPortal(plr.mapID, townInst.fieldID, doorPos))
	}

	if plr.party != nil {
		ownerIdx := plr.party.getPlayerIndex(plr.ID)
		for _, viewer := range townInst.players {
			if viewer == nil || viewer.party == nil || viewer.party.ID != plr.party.ID {
				continue
			}
			viewer.Send(packetMapPortalParty(ownerIdx, plr.mapID, townInst.fieldID, doorPos))
		}
		plr.Send(packetMapPortalParty(ownerIdx, plr.mapID, townInst.fieldID, doorPos))
	} else {
		plr.Send(packetMapPortal(plr.mapID, townInst.fieldID, doorPos))
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
				if plr.townPortalIndex >= 0 && plr.townPortalIndex < len(townInst.portals) {
					townInst.portals[plr.townPortalIndex].resetTownPortal()
				}
				removeDoorFromInstance(townInst, plr.townDoorSpawnID, -1, plr.ID)
			}
		}
	}

	// Reset player-side door state
	plr.resetDoorInfo()
}

// removeDoorFromInstance removes the door object and optional portal from an instance and broadcasts removal
func removeDoorFromInstance(inst *fieldInstance, spawnID int32, portalIndex int, ownerID int32) {
	inst.send(packetMapRemoveMysticDoor(spawnID, false))

	if portalIndex >= 0 {
		inst.removePortalAtIndex(portalIndex)
	}

	removePkt := packetMapRemovePortal()
	inst.send(removePkt)

	delete(inst.mysticDoors, ownerID)

	for _, p := range inst.players {
		if p != nil && p.ID == ownerID {
			p.resetDoorInfo()
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
					if doorInfo.portalIndex >= 0 && doorInfo.portalIndex < len(townInst.portals) {
						townInst.portals[doorInfo.portalIndex].resetTownPortal()
					}
					removeDoorFromInstance(townInst, doorInfo.spawnID, -1, ownerID)
				}
			}
		}
	}
}
