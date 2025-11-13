npc.sendNext("How did you go through such a treacherous road to get here? Incredible! #bThe Breath of Lava#k is here. Please give this to my brother. You'll be meeting the one you've been looking for very soon.")

var itemBreathOfLava = 4031062
var returnMap = 211042300

if (plr.itemCount(itemBreathOfLava) >= 1) {
    npc.sendOk("You already have #bThe Breath of Lava#k. I'll send you back now.")
    plr.warp(returnMap)
} else {
    if (!plr.giveItem(itemBreathOfLava, 1)) {
        npc.sendBackNext("Your Etc. inventory seems to be full. Please make room in order to receive the item.", false, true)
    } else {
        plr.giveEXP(15000)
        plr.warp(returnMap)
    }
}