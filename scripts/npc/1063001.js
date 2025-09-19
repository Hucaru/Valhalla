var pos = plr.position();

if (!pos || pos.y > -1755) {
    npc.sendOk("You can't see the inside of the pile of flowers very well because you're too far. Go a little closer.");
} else if (!plr.giveItem(4031026, 20)) {
    npc.sendOk("Etc item inventory is full.");
} else {
    plr.warp(105000000);
}
