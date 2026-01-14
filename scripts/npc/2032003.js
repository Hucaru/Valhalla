npc.sendNext("How did you go through such a treacherous road to get here? Incredible! #bThe Breath of Lava#k is here. Please give this to my brother. You'll be meeting the one you've been looking for very soon.")

var itemBreathOfLava = 4031062
var returnMap = 211042300
var questStage2 = 7001            // Quest ID for Stage 2 completion tracking
var questComplete = "end"         // Quest data value for completed stages

if (plr.itemCount(itemBreathOfLava) >= 1) {
    npc.sendOk("You already have #t" + itemBreathOfLava + "#. I'll send you back now.")
    plr.warp(returnMap)
} else {
    plr.giveEXP(15000)
    plr.giveItem(itemBreathOfLava, 1)
    // Mark Stage 2 as completed
    plr.setQuestData(questStage2, questComplete)
    npc.sendOk("Take this and head back safely. You have completed Stage 2!")
    plr.warp(returnMap)
}