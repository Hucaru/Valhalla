// Exit NPC - LudiPQ (party2_out)
// Soldier Anderson - Allows players to exit the PQ

var pass = 4001022;
var key = 4001023;

// Clean up quest items
if (plr.itemCount(pass) > 0) {
    plr.removeItemsByID(pass, plr.itemCount(pass));
}
if (plr.itemCount(key) > 0) {
    plr.removeItemsByID(key, plr.itemCount(key));
}

// Offer to exit
if (npc.sendYesNo("You'll have to start over from scratch if you want to take a crack at this quest after leaving this stage. Are you sure you want to leave this map?")) {
    plr.warp(221024500); // Exit to Ludibrium
}
