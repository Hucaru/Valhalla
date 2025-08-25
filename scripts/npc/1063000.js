if (plr.getPosition().y > -2205) {
    npc.sendOk("You can't see the inside of the pile of flowers very well because you're too far. Go a little closer.")
} else if (plr.getInventoryFreeSlot(4) < 1) {
    npc.sendOk("Etc item inventory is full.")
} else {
    plr.giveItem(4031025, 10)
    plr.warp(105000000)
}

// Generate by kimi-k2-instruct