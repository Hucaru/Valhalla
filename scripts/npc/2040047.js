// Exit NPC - LudiPQ (party2_out)
// Soldier Anderson - Allows players to exit the PQ

var pass = 4001022;
var key = 4001023;

var isExitRoom = (plr.mapID() === 922010000);

var prompt = isExitRoom
    ? "All done? See you next time!\nWould you like to leave?"
    : "You'll have to start over from scratch if you want to take a crack at this quest after leaving this stage.\nAre you sure you want to leave this map?";

if (npc.sendYesNo(prompt)) {
    if (isExitRoom) {
        plr.warp(221024500);
    } else {
        if (plr.itemCount(pass) > 0) {
            plr.removeItemsByID(pass, plr.itemCount(pass));
        }
        if (plr.itemCount(key) > 0) {
            plr.removeItemsByID(key, plr.itemCount(key));
        }
        plr.warp(922010000);
    }
}
