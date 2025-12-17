// Nella

// TODO: Flavour text

if (plr.mapID() == 103000890) {
    if (npc.sendYesNo("Exit to kerning city?")) {
        plr.warp(103000000);
    }
} else {
    if (npc.sendYesNo("You'll have to start over from scratch if you want to take a crack at this quest after leaving this stage. Are you sure you want to leave this map?")) {
        plr.leavePartyQuest()
    }
}