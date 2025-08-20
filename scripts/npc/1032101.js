if (plr.getQuestStatus(21714) == 1 || plr.getQuestStatus(21717) == 1 || plr.getQuestStatus(21718) == 1) {
    npc.sendOk("You will be moved to the South Secret Forest.")
    plr.warp(910100002)
} else {
    npc.sendOk("You can't go to the Secret Forest anytime you want.")
}
// Generate by kimi-k2-instruct