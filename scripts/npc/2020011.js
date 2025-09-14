if (plr.level() < 50) {
    npc.sendOk("You're nowhere near ready to fight Zakum. I wouldn't suggest going in there until you're at least level 50.");
} else if (Math.floor(plr.job() / 100 % 10) !== 4) {
    npc.sendBackNext("You're no thief. I am not qualified to judge you. If you want to explore Zakum, you will need to find a master of your job class to be your guide.", true, true);
} else {
    npc.sendBackNext("You should be able to stand against Zakum. Find #b#p2030008##k deep within the Dead Mine. I will allow it.", false, true);
    npc.sendBackNext("Then I will send you to #bThe Door to Zakum#k, where #b#p2030008##k is.", true, true);
    plr.warp(211042300);
}