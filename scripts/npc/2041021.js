npc.sendSelection("#e<Boss: Populatus>#n \r\nThat Troublemaker Papulatus keeps causing all kinds of dimensional cracks. He must be stopped! Can you help? \r\n\r\n#L0#Easy Mode (Level 115 and above)#l")
var sel = npc.selection()

if (sel == 0) {
    if (!plr.isPartyLeader()) {
        npc.sendOk("Only party leaders can initiate entry.")
    } else if (plr.itemQuantity(4031179) < 1) {
        npc.sendOk("You don't have the item needed to meet Papulatus. I'll give you what I happened to have on me. \r\n\r\n#e<Required Items>#n #v4031179# Piece of Cracked Dimension")
    } else {
        var members = plr.getPartyMembers()
        var allOk = true
        for (var i = 0; i < members.length; i++) {
            if (members[i].level() < 115) {
                allOk = false
                break
            }
        }
        if (!allOk) {
            npc.sendOk("All party members must be at least level 115.")
        } else {
            var em = npc.getEventManager("Populatus")
            var prop = em.getProperty("state")
            if (prop == null || prop == 0) {
                em.startInstance(plr.getParty(), plr.getMap(), 200)
            } else {
                npc.sendNext("Another party is already inside.")
            }
        }
    }
}

// Generate by kimi-k2-instruct