if (plr.position.x < -50 || plr.position.x > 250 || plr.position.y > 600) {
    npc.sendOk("You cannot see very well because you're too far. Go a little closer.");

}

if (plr.itemCount(4031039) > 0) {
    npc.sendOk("Looking carefully into #p1052008#, there seems to be a shiny object inside but since your etc. inventory is full, that item is unattainable.");

}

plr.giveItem(4031039, 1);
npc.sendBackNext("Looking carefully into #p1052008#, there seems to be a shiny object inside. Reached out with a hand and was able to attain a small coin.", false, true);

plr.warp(103020000, 0);