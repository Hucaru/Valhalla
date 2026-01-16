// Stage 9 NPC - LudiPQ (Boss Stage)
var key = 4001023;
var props = map.properties();

if (!plr.isPartyLeader()) {
    npc.sendOk("Here is the information about the 9th stage. Now is your chance to finally get your hands on the real culprit. Go right and you'll see a monster. Defeat it to find a monstrous #b#o9300012##k appearing out of nowhere. Your task is to defeat him, collect the #b#t4001023##k he has and bring it to me.");
} else if (plr.itemCount(key) >= 1) {
    if (plr.eventMembersOnMap(plr.mapID())) {
        if (npc.sendYesNo("Incredible! You defeated Alishar and obtained the #t" + key + "#! Would you like to proceed to the bonus stage?")) {
            plr.removeItemsByID(key, 1);
            props.clear = true;
            map.showEffect("quest/party/clear");
            map.playSound("Party1/Clear");
            map.portalEffect("gate");
            plr.partyGiveExp(27200);
            plr.warpEventMembers(922011000);
        }
    } else {
        npc.sendOk("Please make sure all members of your party are in the current map.")
    }

} else {
    npc.sendOk("Defeat Alishar and bring me the #b#t" + key + "##k to proceed!");
}
