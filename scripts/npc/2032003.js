npc.sendBackNext("How did you go through such a treacherous road to get here? Incredible! #bThe Breath of Lava#k is here. Please give this to my brother. You'll be meeting the one you've been looking for very soon.", false, true)

var itemBreathOfLava = 4031062
var returnMap = 211042300

if (plr.itemCount(itemBreathOfLava) >= 1) {
    npc.sendOk("You already have #t" + itemBreathOfLava + "#. I'll send you back now.")
    plr.warp(returnMap)
} else {
    plr.giveEXP(15000)
    plr.giveItem(itemBreathOfLava, 1)
    npc.sendOk("Take this and head back safely.")
    plr.warp(returnMap)
}