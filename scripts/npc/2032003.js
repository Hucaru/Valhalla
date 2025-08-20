npc.sendBackNext("How did you go through-such treacherous road to get here?? Incredible! #bThe Breath of Lava#k is here. Please give this to my brother. You'll finally be meeting up with the one you've been looking for, very soon.", false, true)

if (plr.giveItem(4031062, 1)) {
    plr.giveExp(15000)
    plr.warp(211042300)
} else {
    npc.sendBackNext("Your etc, inventory seems to be full. Please make room in order to receive the item.", true, true)
}

// Generate by kimi-k2-instruct