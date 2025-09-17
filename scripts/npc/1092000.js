if (plr.getQuestStatus(2905) !== 1) {
    npc.sendOk("Do you want to make some delicious dishes for the crew of the Nautilus? I can teach you how.");
} else {
    plr.warp(912000100);
}