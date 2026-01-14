// Stage 1 NPC - LudiPQ
var pass = 4001022;
var props = map.properties();

if (!plr.isPartyLeader()) {
    npc.sendOk("Here is information about the 1st stage. You'll see monsters at different points on the map. These monsters have an item called #b#t4001022##k, which opens the door to another dimension. Defeat the monsters, collect #b25 #t4001022#s#k and give them to your party leader, who will in turn give them to me.");
} else {
    if (plr.itemCount(pass) >= 25) {
        if (npc.sendYesNo("Good job! You have collected 25 #t" + pass + "#s. Would you like to move to the next stage?")) {
            plr.removeItemsByID(pass, 25);
            props.clear = true;
            map.showEffect("quest/party/clear");
            map.playSound("Party1/Clear");
            map.portalEffect("gate");
            npc.sendOk("The portal to the next stage is now open!");
        }
    } else {
        npc.sendOk("You need to collect #b25 #t" + pass + "#s#k. You currently have #b" + plr.itemCount(pass) + "#k.");
    }
}
