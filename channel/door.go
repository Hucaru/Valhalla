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
		destName:    "sp",
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
				doorInst.send(packetMapRemoveMysticDoor(plr.doorSpawnID, true))
				doorInst.removePortalAtIndex(plr.doorPortalIndex)

				removePkt := packetMapRemovePortal()
				for _, p := range doorInst.players {
					if p == nil {
						continue
					}
					authorized := (p.ID == plr.ID)
					if !authorized && plr.party != nil && p.party != nil && p.party.ID == plr.party.ID {
						authorized = true
					}
					if authorized {
						p.Send(removePkt)
					}
				}

				delete(doorInst.mysticDoors, plr.ID)
			}
		}
	}

	if plr.townDoorMapID != 0 {
		if townField, ok := plr.inst.server.fields[plr.townDoorMapID]; ok {
			if townInst, err := townField.getInstance(0); err == nil {
				townInst.send(packetMapRemoveMysticDoor(plr.townDoorSpawnID, true))
				removePkt := packetMapRemovePortal()
				for _, p := range townInst.players {
					if p == nil {
						continue
					}
					authorized := (p.ID == plr.ID)
					if !authorized && plr.party != nil && p.party != nil && p.party.ID == plr.party.ID {
						authorized = true
					}
					if authorized {
						p.Send(removePkt)
					}
				}
				delete(townInst.mysticDoors, plr.ID)
			}
		}
	}

	plr.doorMapID = 0
	plr.doorSpawnID = 0
	plr.doorPortalIndex = 0
	plr.townDoorMapID = 0
	plr.townDoorSpawnID = 0
	plr.townPortalIndex = 0
}

// mysticDoorExpired handles door expiration
func mysticDoorExpired(playerID, sourceMapID, townMapID int32, server *Server) {
	var plr *Player

	if sourceField, ok := server.fields[sourceMapID]; ok {
		if sourceInst, err := sourceField.getInstance(0); err == nil {
			for _, p := range sourceInst.players {
				if p.ID == playerID {
					plr = p
					break
				}
			}
		}
	}

	if plr == nil && townMapID > 0 {
		if townField, ok := server.fields[townMapID]; ok {
			if townInst, err := townField.getInstance(0); err == nil {
				for _, p := range townInst.players {
					if p.ID == playerID {
						plr = p
						break
					}
				}
			}
		}
	}

	if plr != nil {
		removeMysticDoor(plr)
	} else {
		if sourceField, ok := server.fields[sourceMapID]; ok {
			if sourceInst, err := sourceField.getInstance(0); err == nil {
				if doorInfo, exists := sourceInst.mysticDoors[playerID]; exists {
					sourceInst.send(packetMapRemoveMysticDoor(playerID, true))
					sourceInst.removePortalAtIndex(doorInfo.portalIndex)
					delete(sourceInst.mysticDoors, playerID)
				}
			}
		}

		if townMapID > 0 {
			if townField, ok := server.fields[townMapID]; ok {
				if townInst, err := townField.getInstance(0); err == nil {
					if _, exists := townInst.mysticDoors[playerID]; exists {
						townInst.send(packetMapRemoveMysticDoor(playerID, true))
						delete(townInst.mysticDoors, playerID)
					}
				}
			}
		}
	}
}
