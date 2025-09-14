if (npc.sendYesNo("Once I lay my hand on the statue, a strange light covers me and it feels like I am being sucked into somewhere else. Will it be okay to go back to #m105000000#?")) {
    plr.warp(105000000);
} else {
    npc.sendOk("Once I took my hand off the statue it got quiet, as if nothing happened.");
}