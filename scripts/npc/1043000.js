var pos = plr.position();

if (!pos || pos.y > -3322) {
    npc.sendOk("You can't see the inside of the pile of flowers very well because you're too far. Go a little closer.");
} else if (plr.itemCount(4031020) >= 1) {
    npc.sendOk("You already have #b#t4031020##k.");
} else if (npc.sendYesNo("Are you sure you want to take #b#t4031020##k with you?")) {
    if (plr.giveItem(4031020, 1)) {
        plr.warp(101000000);
    } else {
        npc.sendOk("Please make room in your Etc inventory.");
    }
}