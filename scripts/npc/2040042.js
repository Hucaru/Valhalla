// Stage 7 NPC - LudiPQ
var pass = 4001022;
var props = map.properties();

if (!plr.isPartyLeader()) {
    npc.sendOk("Here is information about the 7th stage. Here you will find a ridiculously powerful monster called #b#o9300010##k. Defeat the monster and get the #b#t4001022##k needed to proceed to the next stage. Please collect #b3 #t4001022#s#k. Defeat it from afar!");
} else {
    if (plr.itemCount(pass) >= 3) {
        if (npc.sendYesNo("Good job! You have collected 3 #t" + pass + "#s. Would you like to move to the next stage?")) {
            plr.removeItemsByID(pass, 3);
            props.clear = true;
            map.showEffect("quest/party/clear");
            map.playSound("Party1/Clear");
            map.portalEffect("gate");
            npc.sendOk("The portal to the next stage is now open!");
        }
    } else {
        npc.sendOk("You need to collect #b3 #t" + pass + "#s#k. You currently have #b" + plr.itemCount(pass) + "#k.");
    }
}
