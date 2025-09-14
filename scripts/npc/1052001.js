/*
** NPC: Dark Lord
** Location: Kerning City — Thieves' Headquarters
*/

if (plr.job() === 0) {
    if (plr.getLevel() >= 10) {
        npc.sendBackNext(
            "Do you want to be a Thief? It is an important and final choice. You will not be able to turn back.",
            false, true
        )
        if (npc.sendYesNo("You look qualified for this. With sharp eyes and cold precision, a Thief strikes from the shadows. Do you want to become a #rThief#k?")) {
            plr.setJob(400)
            plr.giveItem(1332063, 1)
            npc.sendBackNext("Alright! You are a Thief from here on out... Here's a little bit of my power to you... Haahhhh!", true, true)
            npc.sendBackNext("I've added a few slots to your Equip and ETC inventories, and your strength has increased. Train hard. I'll be watching your every move, so don't let me down.", true, true)
            npc.sendOk("I also gave you a little bit of #bSP#k. Open the #bSkill Menu#k to learn skills. Some skills require others first, so choose wisely.")
        } else {
            npc.sendOk("Come back once you have thought about it some more.")
        }
    } else {
        npc.sendOk("Train a bit more until you reach #bLevel 10#k and I can show you the way of the #rThief#k.")
    }

} else if (plr.job() === 400) {
    if (plr.getLevel() >= 30) {
        npc.sendBackNext("Hmmm... you seem to have gotten a whole lot stronger. Ready to take the next step?", false, true)

        var expl = npc.sendMenu(
            "When you are ready, choose your path.",
            "Please explain the role of the Assassin.",
            "Please explain the role of the Bandit.",
            "I'll choose my occupation!"
        )
        if (expl === 0) {
            npc.sendOk("Assassins excel at ranged throwing weapons, striking swiftly from the dark with deadly precision.")
        } else if (expl === 1) {
            npc.sendOk("Bandits fight up close with daggers and guile, shattering foes with rapid, ruthless strikes.")
        }

        var chosen = -1
        while (chosen === -1) {
            var branch = npc.sendMenu(
                "Hmmm, have you made up your mind? Choose the 2nd job advancement of your liking...",
                "Assassin",
                "Bandit"
            )
            var jobName = (branch === 0) ? "Assassin" : "Bandit"
            var jobId   = (branch === 0) ? 410       : 420

            if (npc.sendYesNo("So you want to advance as a #b" + jobName + "#k? Once you decide, you can't go back. Are you sure?")) {
                plr.setJob(jobId)
                if (branch === 0) {
                    npc.sendBackNext("Alright, from here on out you are the #bAssassin#k. Keep training and hone your skills in the shadows.", true, true)
                } else {
                    npc.sendBackNext("Alright, from here on out you are the #bBandit#k. Keep training and hone your skills in the shadows.", true, true)
                }
                npc.sendBackNext("Your #bUse#k and #bETC#k inventories have been expanded. Your Max HP and MP have also increased. Go check it out!", true, true)
                npc.sendOk("I have also given you a little bit of #bSP#k. Open the #bSkill Menu#k to enhance your 2nd job skills. Some skills require others first; remember that.")
                chosen = branch
            } else {
                npc.sendOk("Take your time. This decision is important.")
            }
        }
    } else {
        npc.sendOk("Keep training as a Thief. Return to me at #rLevel 30#k for your next advancement.")
    }

} else if (plr.job() === 410 || plr.job() === 420) {
    npc.sendOk("Walk the path you've chosen with pride. Keep training and growing stronger.")

} else {
    npc.sendOk("To those that want to be a Thief, come...")
}