// Check level first
if (plr.getLevel() < 65) {
    npc.sendOk("The Balrog imprisoned here is incredibly dangerous. Only a fearsome warrior can join the #e<Boss: Balrog>#n expedition. \r\nYou must be above Lv. 65 to join.")
}

// Present boss entry option
npc.sendSelection(
    "#e<Boss: Balrog>#n \r\n" +
    "The Balrog is locked up over yonder. It is a lord of darkness that was sealed by Master Manji and the hero Tristan a long time ago. But since it recently began to gain back its power, the seal has grown unstable. It needs to be reinforced as quickly as possible... \r\n" +
    "(All Channels / Lv. 65 and above / 1 - 6 players) \r\n\r\n" +
    "#L0##bRequest to enter <Boss: Balrog>."
)
var sel = npc.selection()

// Only proceed if option 0 was clicked
if (sel === 0) {
    // Check if party leader
    if (!plr.isPartyLeader()) {
        npc.sendOk("Only party leaders can initiate entry.")
    }

    // Check all party members on same map
    if (plr.partyMembersOnMapCount() !== plr.getPartySize()) {
        npc.sendNext("All party members must be on the same map to enter.")
    }

    // Check all party members level 65+
    for (var i = 0; i < plr.getPartySize(); i++) {
        if (plr.getPartyMember(i).getLevel() < 65) {
            npc.sendOk("Party members must be at least Lv. 65.")
        }
    }

    // Check if instance is available
    var em = npc.getEventManager("BossBalrog_EASY")
    if (em && (em.getProperty("state") == null || em.getProperty("state") == 0)) {
        em.startInstance(plr.getParty(), plr.getMap(), 200)
    } else {
        npc.sendNext("Another party is already inside.")
    }
}