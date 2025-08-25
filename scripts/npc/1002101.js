if (plr.getQuestStatus(22579) < 2) {
    npc.sendOk("I'm just a retired crewman. I'm focused on training powerful Explorers now.")
} else {
    if (npc.sendYesNo("Do you want to go to the island in John's Map right now?")) {
        plr.warp(200090080)
        plr.startMapTimeLimitTask(30, 914100000)
    } else {
        npc.sendOk("Ah, you still have businesses left in Lith Harbor.")
    }
}

// Generate by kimi-k2-instruct