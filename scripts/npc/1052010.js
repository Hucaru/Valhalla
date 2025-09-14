if (plr.getPosition().x < -50 || plr.getPosition().x > 250 || plr.getPosition().y > 600) {
    npc.sendOk("You cannot see very well because you're too far. Go a little closer.");
} else if (plr.freeSlots(4) < 1) {   // ETC
    npc.sendBackNext("Looking carefully into Treasure Chest there seems to be a shiny object inside but since your etc. inventory is full, that item is unattainable.", true, true);
    plr.warp(103020000);
} else {
    var got = plr.giveItem(4031041, plr.itemCount(4031041) ? 0 : 1);
    npc.sendBackNext("Looking carefully into Treasure Chest there seems to be a sack of something that contains shiny object. Reached out with a hand and was able to attain a heavy sack of coins.", true, true);
    plr.warp(103020000);
}