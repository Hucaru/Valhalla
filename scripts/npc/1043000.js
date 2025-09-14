if (plr.getPosition().y > -3322) {
    npc.sendOk("You can't see the inside of the pile of flowers very well because you're too far. Go a little closer.");
}

if (plr.itemCount(4031020) >= 1) {
    npc.sendOk("Etc item inventory is full.");
}

if (npc.sendYesNo("Are you sure you want to take #b#t4031020##k with you?")) {
    plr.giveItem(4031020, 1);
    plr.warp(101000000);
}