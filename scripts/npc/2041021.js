// Boss: Populatus dialog
npc.sendSelection("#e<Boss: Populatus>#n \r\nThat Troublemaker Papulatus keeps causing all kinds of dimensional cracks. He must be stopped! Can you help? \r\n\r\n#L0#Easy Mode (Level 115 and above)");

// Perform party & requirement checks
if (!plr.inParty()) {
    npc.sendOk("Only party leaders can initiate entry.");

}

if (!plr.isPartyLeader()) {
    npc.sendOk("Only party leaders can initiate entry.");

}

if (plr.itemCount(4031179) < 1) {
    npc.sendOk("You don't have the item needed to meet Papulatus. I'll give you what I happened to have on me. \r\n\r\n#e<Required Items>#n #v4031179# Piece of Cracked Dimension");

}

for (var i = 0; i < plr.partyMembersOnMapCount(); i++) {
    if (plr.getPartyMemberLevel(i) < 115) {
        npc.sendOk("All party members must be at least level 115.");

    }
}

// Attempt to start boss instance
var prop = instanceProperties();
if (!prop.state || prop.state == 0) {
    // Start the boss instance
    for (var i = 0; i < plr.partyMembersOnMapCount(); i++) {
        instanceProperties().addPartyMember(plr.getPartyMemberId(i));
    }
    instanceProperties().state = 1;
    instanceProperties().eventId = "Populatus";
} else {
    npc.sendNext("Another party is already inside.");
}