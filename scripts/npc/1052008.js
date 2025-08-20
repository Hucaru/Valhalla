if (plr.getPosition().x < -50 || plr.getPosition().x > 250 || plr.getPosition().y > 600) {
    npc.sendOk("You cannot see very well because you're too far. Go a little closer.")
} else if (plr.getInventoryFreeSlot(4) < 1) {
    npc.sendNext("Looking carefully into #p1052008#, there seems to be a shiny object inside but since your etc. inventory is full, that item is unattainable.")
} else {
    plr.giveItem(4031039, plr.itemQuantity(4031039) ? 0 : 1)
    npc.sendNext("Looking carefully into #p1052008#, there seems to be a shiny object inside. Reached out with a hand and was able to attain a small coin.")
    plr.warp(103020000)
}

// Generate by kimi-k2-instruct