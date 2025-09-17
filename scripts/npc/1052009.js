// Check if player is too far away
if (plr.getX() < -50 || plr.getX() > 250 || plr.getY() > 600) {
    npc.sendOk("You cannot see very well because you're too far. Go a little closer.");
}

// Check etc inventory space
if (plr.itemCount(4031040) === 0) {
    if (npc.sendNext("Looking carefully into Treasure Box, there seems to be a shiny object inside but since your etc. inventory is full, that item is unattainable.") &&
        plr.getMapId() === 910360102) {
        plr.warp(103020000);
    }
} else {
    if (npc.sendNext("Looking carefully into Treasure Box, there seems to be a stack of papers in there. I reached out my hand and voila, a huge stack of money.")) {
        plr.giveItem(4031040, 0); // keep quantity
        plr.warp(103020000);
    }
}