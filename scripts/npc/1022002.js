if (npc.sendYesNo("What? You want a shot at sealing Balrog? A weakling like you might not make it back in one piece! Well, I suppose it isn't my business. Alright, you need to pay a fee of #b10,000 Mesos#k. Do you have enough Mesos on you?")) {
    if (plr.level() < 45) {
        npc.sendOk("Rookies like yourself under LV. 45 don't have any right to go there at all. Now shoo!")
    } else {
        if (plr.mesos() < 10000) {
            npc.sendOk("You don't have enough Mesos. How dare you even dream of participating without the right amount of Mesos?! Scram!")
        } else {
            npc.sendBackNext("Alright, don't disappoint me now. You'll be able to participate in the Expedition Team if you visit my apprentice #b#p1061014##k, upon your arrival.", true, true)
            plr.takeMesos(10000)
            plr.updateInfoQuest(1022002, 1)
            plr.warp(105100100)
        }
    }
} else {
    npc.sendNext("Ah, you're aware of how precious life is, don't you? Stop wasting my time and leave.")
}

// Generate by kimi-k2-instruct