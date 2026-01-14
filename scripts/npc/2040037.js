// Stage 2 NPC - LudiPQ
var pass = 4001022;
var props = map.properties();

if (!plr.isPartyLeader()) {
    npc.sendOk("Here is information about the 2nd stage. You'll see crates all over the map. Break a box and you will be sent to another map or rewarded with a #t4001022#. Search each box, collect #b15 #t4001022#s#k and bring them to your party leader.");
} else {
    if (plr.itemCount(pass) >= 15) {
        if (npc.sendYesNo("Good job! You have collected 15 #t" + pass + "#s. Would you like to move to the next stage?")) {
            plr.removeItemsByID(pass, 15);
            props.clear = true;
            map.showEffect("quest/party/clear");
            map.playSound("Party1/Clear");
            map.portalEffect("gate");
            npc.sendOk("The portal to the next stage is now open!");
        }
    } else {
        npc.sendOk("You need to collect #b15 #t" + pass + "#s#k. You currently have #b" + plr.itemCount(pass) + "#k.");
    }
}
