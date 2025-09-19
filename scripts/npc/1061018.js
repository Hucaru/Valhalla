if (npc.sendYesNo("Would you like to forfeit and exit?")) {
    plr.warp(105100100);
} else {
    npc.sendNext("Try a bit harder.");
}
