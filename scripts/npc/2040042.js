// Welcome to the fourth stage
npc.sendBackNext("Welcome to the fourth stage. Here, you must face the powerful #b#o9300010##k. #b#o9300010##k is a fearsome opponent, so do not let your guard down. Once you defeat it, let me know and I'll show you to the next stage.", false, true)

// Check if stage4 is cleared and leader
var eim = plr.instanceProperties()
if (eim.stage4 > 0) {
    npc.sendBackNext("Congratulations on clearing the quests for this stage. Please use the portal you see over there and move on to the next stage.", true, true)
} else {
    npc.sendBackNext("Wow, not a single #b#o9300010##k left! I'm impressed! I can open the portal to the next stage now.", false, true)
    eim.stage4 = 1
    npc.sendBackNext("The portal that leads you to the next stage is now open.", true, true)
}

// Generate by kimi-k2-instruct