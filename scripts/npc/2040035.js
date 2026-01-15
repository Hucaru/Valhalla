// Bonus Exit NPC - LudiPQ

var pass = 4001022;
var key = 4001023;

if (npc.sendYesNo("Did you have a fun time in the bonus map? Let's get you back to the starting area.")) {
    // Clean up quest items
    if (plr.itemCount(pass) > 0) {
        plr.removeItemsByID(pass, plr.itemCount(pass));
    }
    if (plr.itemCount(key) > 0) {
        plr.removeItemsByID(key, plr.itemCount(key));
    }
    plr.warp(221024500); // Exit to Ludibrium
}
