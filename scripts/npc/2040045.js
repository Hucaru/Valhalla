// Bonus Exit NPC - LudiPQ
var pass = 4001022;
var key = 4001023;

if (npc.sendYesNo("Congratulations! You defeated Alishar. \nWelcome to the Bonus Stage, break the boxes here for extra prizes. \nDo you want to leave the Bonus Stage now?")) {
    // Clean up quest items
    if (plr.itemCount(pass) > 0) {
        plr.removeItemsByID(pass, plr.itemCount(pass));
    }
    if (plr.itemCount(key) > 0) {
        plr.removeItemsByID(key, plr.itemCount(key));
    }

    plr.warp(221024500);
}
