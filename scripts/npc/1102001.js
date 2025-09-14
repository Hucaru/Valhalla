npc.sendYesNo("Have you found all the proof for the test? Do you want to get out of here?");
if (npc.selection() === 1) {
    plr.warp(130020000);
} else {
    npc.sendOk("Take your time.");
}