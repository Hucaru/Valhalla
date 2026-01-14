// Stage 3 NPC - LudiPQ  
var pass = 4001022;
var props = map.properties();

if (!plr.isPartyLeader()) {
    npc.sendOk("Here is information about the 3rd stage. Here you will see monsters and boxes. If you defeat the monsters or break boxes, they will drop #b#t4001022##k. The number you need to collect is determined by the answer to a question (HP of Level 1 - Min level for magician - Min level for rogue = 32).");
} else {
    if (plr.itemCount(pass) >= 32) {
        if (npc.sendYesNo("Good job! You have collected 32 #t" + pass + "#s. Would you like to move to the next stage?")) {
            plr.removeItemsByID(pass, 32);
            props.clear = true;
            map.showEffect("quest/party/clear");
            map.playSound("Party1/Clear");
            map.portalEffect("gate");
            npc.sendOk("The portal to the next stage is now open!");
        }
    } else {
        npc.sendOk("You need to collect #b32 #t" + pass + "#s#k. You currently have #b" + plr.itemCount(pass) + "#k.");
    }
}
