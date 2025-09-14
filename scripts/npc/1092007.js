// 盲俠
if (plr.getQuestStatus(2925) !== 1) {
    npc.sendOk("Hey, this isn't a zoo! You got business in there?");
} else {
    plr.warp(912040300);
}