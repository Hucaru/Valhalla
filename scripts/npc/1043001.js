// Herb patch handler
if (plr.getPosition().y > -2962) {
    npc.sendOk("You can't see the inside of the pile of flowers very well because you're too far. Go a little closer.");
} else if (!giveItem(4031032, 1)) {
    npc.sendOk("Etc item inventory is full.");
} else {
    if (npc.sendYesNo("Are you sure you want to take #b#t4031032##k with you?")) {
        plr.giveItem(4031032, 1);
        plr.warp(101000000);
    }
}