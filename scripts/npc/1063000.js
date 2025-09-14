if (plr.getPosition().y > -2205) {
    npc.sendOk("You can't see the inside of the pile of flowers very well because you're too far. Go a little closer.");
} else if (!npc.giveItem(4031025, 10)) {
    npc.sendOk("Etc item inventory is full.");
} else {
    npc.sendOk("You received #v4031025# x10!");
    plr.warp(105000000);
}