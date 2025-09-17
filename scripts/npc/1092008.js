if (plr.getQuestStatus(2915) == 1) {
    plr.warp(912040100, 1);
} else if (plr.getQuestStatus(2916) == 1) {
    plr.warp(912040200, 1);
} else {
    npc.sendOk("The Training Room is off-limits unless you are scheduled.");
}