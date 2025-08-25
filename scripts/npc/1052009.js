// Check distance
if (plr.getPosition().x < -50 || plr.getPosition().x > 250 || plr.getPosition().y > 600) {
    npc.sendOk("You cannot see very well because you're too far. Go a little closer.")
} else if (plr.getInventorySlotCount(4) - plr.getInventoryItemCount(4) < 1) {
    npc.sendNext("Looking carefully into Treasure Box, there seems to be a shiny object inside but since your etc. inventory is full, that item is unattainable.")
} else {
    plr.giveItem(4031040, plr.itemQuantity(4031040) ? 0 : 1)
    npc.sendNext("Looking carefully into Treasure Box, there seems to be a stack of papers in there. I reached out my hand and voila, a huge stack of money.")
    plr.warp(103020000, 0)
}

// Generate by kimi-k2-instruct