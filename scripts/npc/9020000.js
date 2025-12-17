// Lakelis

// TODO: Flavour text

var maps = [103000800, 103000801, 103000802, 103000803, 103000804, 103000805]

if (plr.isPartyLeader()) {
    var plrs = plr.partyMembersOnMap()

    var badLevel = false

    for (let i = 0; i < plrs.length; i++) {
        if (plrs[i].level() > 30 || plrs[i].level() < 21) {
            badLevel = true
            break
        }
    }

    if (plr.partyMembersOnMapCount() != 4) {
        npc.sendOk("You need to be a party of 4 on the same map")
    } else if (badLevel) {
        npc.sendOk("Someone in your party is not the correct level")
    } else {
        for (let instance = 0; instance < 1; instance++) {
            var count = 0;

            for(let i = 0; i < maps.length; i++) {
                var m = map.getMap(maps[i], instance)
                count += m.playerCount()
            }

            if (count == 0) {
                plr.startPartyQuest("kerning_pq", instance)
            } else {
                npc.sendOk("A party is already doing the quest, please come back another time")
            }
        }
    }
} else {
    npc.sendOk("You need to be party leader to start a party quest")
}