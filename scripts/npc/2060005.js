// Check quest status
if (plr.getQuestStatus(6002) > 1) {
    npc.sendOk("You already guarded the pig, and you did just fine.")
} else if (plr.getQuestStatus(6002) < 1) {
    npc.sendOk("What pig? Where did you hear about that?")
} else if (plr.itemQuantity(4031508) > 5 && plr.itemQuantity(4031507) > 5) {
    npc.sendOk("I don't need another one of #bKenta's Reports#k and I'm all stocked up on #bPheromone#k. You don't need to go in.")
} else {
    // Attempt to start the event
    var em = npc.getEventManager("q6002")
    var prop = em.getProperty("state")
    if (prop == null || prop == 0) {
        plr.takeItem(4031507, plr.itemQuantity(4031507))
        plr.takeItem(4031508, plr.itemQuantity(4031508))
        em.startInstance(plr)
    } else {
        npc.sendOk("Someone is attempting to protect the Watch Hog already. Please try again later.")
    }
}

// Generate by kimi-k2-instruct