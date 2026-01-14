// Stage 5 NPC - LudiPQ
var pass = 4001022;
var props = map.properties();

if (!plr.isPartyLeader()) {
    npc.sendOk("Here is information about the 5th stage. Here you will find many spaces with monsters. Your duty is to collect #b24 #t4001022#s#k. There will be cases where you need to be of a certain profession. There is a monster called #b#o9300013##k that only rogues can pass. There is also a route that only mages can take.");
} else {
    if (plr.itemCount(pass) >= 24) {
        if (npc.sendYesNo("Good job! You have collected 24 #t" + pass + "#s. Would you like to move to the next stage?")) {
            plr.removeItemsByID(pass, 24);
            props.clear = true;
            map.showEffect("quest/party/clear");
            map.playSound("Party1/Clear");
            map.portalEffect("gate");
            npc.sendOk("The portal to the next stage is now open!");
        }
    } else {
        npc.sendOk("You need to collect #b24 #t" + pass + "#s#k. You currently have #b" + plr.itemCount(pass) + "#k.");
    }
}
