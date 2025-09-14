// Dolphin Taxi - Sharp Unknown / Herb Town / Aquarium routing

if (plr.getMapId() === 230000000) {
    // In Aqua Road
    var sel = npc.askSelection("Oceans are all connected to each other. Places you can't reach by foot can be easily reached oversea. How about taking #bDolphin Taxi#k with us today? \r\n\r\n#b#L0#Go to the Sharp Unknown. (Towards Ludibrium/Korean Folk Town)#l\r\n#L1#Go to Herb Town.#l");
    
    if (sel === 0) {
        if (npc.askYesNo("The fare is 1000 mesos. Shall we go?")) {
            if (plr.mesos() < 1000) {
                npc.sendOk("You don't have enough mesos.");
            } else {
                plr.giveMesos(-1000);
                plr.warp(230030200);
            }
        } else {
            npc.sendOk("Okay, next time.");
        }
    } else if (sel === 1) {
        if (npc.askYesNo("The fare is 10000 mesos. Shall we go?")) {
            if (plr.mesos() < 10000) {
                npc.sendOk("You don't have enough mesos.");
            } else {
                plr.giveMesos(-10000);
                plr.warp(251000100);
            }
        } else {
            npc.sendOk("Okay, next time.");
        }
    }
} else if (plr.getMapId() === 251000100) {
    // In Herb Town
    if (npc.askYesNo("Oceans are all connected to each other. Places you can't reach by foot can be easily reached oversea. How about taking #bDolphin Taxi#k with us today? \r\n\r\nThe fare is 10000 mesos. Shall we go to Acuariurm?")) {
        if (plr.mesos() < 10000) {
            npc.sendOk("You don't have enough mesos.");
        } else {
            plr.giveMesos(-10000);
            plr.warp(230000000);
        }
    } else {
        npc.sendOk("Okay, next time.");
    }
}