// Stage 4 NPC - LudiPQ
var pass = 4001022;
var props = map.properties();

if (!plr.isPartyLeader()) {
    npc.sendOk("Here is information about the 4th stage. Here you will find a black space created by the dimensional rift. Inside, you'll find a monster called #b#o9300008##k hiding in the darkness. Defeat the monsters and collect #b6 #t4001022#s#k.");
} else {
    if (plr.itemCount(pass) >= 6) {
        if (npc.sendYesNo("Good job! You have collected 6 #t" + pass + "#s. Would you like to move to the next stage?")) {
            plr.removeItemsByID(pass, 6);
            props.clear = true;
            map.showEffect("quest/party/clear");
            map.playSound("Party1/Clear");
            map.portalEffect("gate");
            npc.sendOk("The portal to the next stage is now open!");
        }
    } else {
        npc.sendOk("You need to collect #b6 #t" + pass + "#s#k. You currently have #b" + plr.itemCount(pass) + "#k.");
    }
}
