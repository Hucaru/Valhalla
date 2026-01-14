// Entry NPC for Ludibrium PQ (party2_enter)
// NPC ID: 2040034

var maps = [922010100, 922010200, 922010300, 922010400, 922010500, 922010600, 922010700, 922010800, 922010900];

if (!plr.isPartyLeader()) {
    npc.sendOk("From this point on, this place is full of dangerous obstacles and monsters. For this reason, I cannot let you go any further. However, if you're interested in saving us and bringing peace back to Ludibrium, that's another story. If you want to defeat a powerful creature that dwells on the summit, please gather your party members. It won't be easy, but... I think you can do it.");
} else {
    var plrs = plr.partyMembersOnMap();
    var badLevel = false;

    for (let i = 0; i < plrs.length; i++) {
        if (plrs[i].level() < 35 || plrs[i].level() > 50) {
            badLevel = true;
            break;
        }
    }

    if (plr.partyMembersOnMapCount() < 3) {
        npc.sendOk("Your party cannot participate in the quest because it does not have 3 members. Please gather 3 people in your party.");
    } else if (badLevel) {
        npc.sendOk("Someone in your party is not between levels 35~50. Please check again.");
    } else {
        var started = false;
        for (let instance = 0; instance < 1; instance++) {
            var count = 0;

            for(let i = 0; i < maps.length; i++) {
                var m = map.getMap(maps[i], instance);
                count += m.playerCount();
            }

            if (count == 0) {
                plr.startPartyQuest("ludibrium_pq", instance);
                started = true;
                break;
            }
        }
        
        if (!started) {
            npc.sendOk("Another party is inside participating in the quest. Please try again after the party opens the vacancy.");
        }
    }
}
