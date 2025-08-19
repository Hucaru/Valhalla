// Athena Pierce

if (plr.job() == 0) {
    if (plr.level() >= 10) {
        npc.sendBackNext("So you decided to become a #rBowman#k?", false, true);
        npc.sendBackNext("It is an important and final choice. You will not be able to turn back.", false, true)

        if (npc.sendYesNo("Do you want to become a #rBowman#k?")) {
            plr.setJob(300)
            plr.giveItem(1452002, 1)
            npc.sendOk("So be it! Now go, and go with pride.")
        }
    } else {
        npc.sendOk("Train a bit more and I can show you the way of the #rBowman#k.")
    }
} else if (plr.job() == 300) {
    npc.sendOk("Not implemented")
} else {
    npc.sendOk("The progress you have made is astonishing.")
}