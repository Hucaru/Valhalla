if (plr.getQuestStatus(2052) !== 1 && plr.getQuestStatus(2053) !== 1 && plr.getQuestStatus(2054) !== 1) {
    npc.sendOk("I laid my hand on the statue but nothing happened.");
} else {
    if (npc.sendYesNo("Once I lay my hand on the statue, a strange light covers me and it feels like I am being sucked into somewhere else. Is it okay to be moved to somewhere else randomly just like that?")) {
        var map;
        if (plr.getQuestStatus(2052) === 1) {
            map = 910530000;
        } else if (plr.getQuestStatus(2053) === 1) {
            map = 910530100;
        } else {
            map = 910530200;
        }
        plr.warp(map);
    } else {
        npc.sendBackNext("Once I took my hand off the statue it got quiet, as if nothing happened.", true, true);
    }
}