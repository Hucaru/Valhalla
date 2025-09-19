var pos = plr.position();

if (!pos || pos.x < -50 || pos.x > 250 || pos.y > 600) {
    npc.sendOk("You cannot see very well because you're too far. Go a little closer.");
} else if (plr.freeSlots(4) < 1) { // ETC inventory tab
    npc.sendBackNext(
        "Looking carefully into Treasure Chest there seems to be a shiny object inside but since your etc. inventory is full, that item is unattainable.",
        true,
        true
    );
    plr.warp(103020000);
} else {
    var alreadyHas = plr.itemCount(4031041) > 0;
    var given = alreadyHas ? true : plr.giveItem(4031041, 1);

    npc.sendBackNext(
        "Looking carefully into Treasure Chest there seems to be a sack of something that contains shiny object. Reached out with a hand and was able to attain a heavy sack of coins.",
        true,
        true
    );
    plr.warp(103020000);
}
