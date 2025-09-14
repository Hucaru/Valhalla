/*
** NPC: Athena Pierce
** Location: Henesys — Bowman Instructional School
*/

if (plr.job() === 0) {
    if (plr.getLevel() >= 10) {
        npc.sendBackNext(
            "Do you want to be a Bowman? It is an important and final choice. You will not be able to turn back.",
            false, true
        )
        if (npc.sendYesNo("You look qualified for this. With keen eyes to spot true threats and the skill to let arrows fly with precision... we need someone like that. Do you want to become a #rBowman#k?")) {
            plr.setJob(300)
            plr.giveItem(1452002, 1)
            npc.sendBackNext("Alright! You are a Bowman from here on out... Here's a little bit of my power to you... Haahhhh!", true, true)
            npc.sendBackNext("I have added slots for your equipment and etc. inventory. You have also gotten much stronger. Train harder, and you may one day reach the very top of the Bowman ranks. I'll be watching you from afar. Please work hard.", true, true)
            npc.sendOk("I also gave you a little bit of #bSP#k. Open the #bSkill Menu#k at the bottom-left to learn skills. Some skills require others first, so choose wisely.")
        } else {
            npc.sendOk("Come back once you have thought about it some more.")
        }
    } else {
        npc.sendOk("Hmm... you are not quite ready yet. Return to me at #bLevel 10#k and I will show you the way of the #rBowman#k.")
    }

} else if (plr.job() === 300) {
    if (plr.getLevel() >= 30) {
        npc.sendBackNext("Well look who's here!... you came back safe! I knew you'd breeze through your early trials. Now, choose your path and I will make you even stronger.", false, true)

        var explain = npc.sendMenu(
            "When you are ready, choose your path.",
            "Please explain the role of the Hunter.",
            "Please explain the role of the Crossbow Man.",
            "I'll choose my occupation!"
        )
        if (explain === 0) {
            npc.sendOk("Hunters are keen and agile archers, striking enemies with rapid arrows and superior precision.")
        } else if (explain === 1) {
            npc.sendOk("Crossbow Men fight tactically with powerful, deliberate shots, exploiting enemy weaknesses at range.")
        }

        var picked = -1
        while (picked === -1) {
            var branch = npc.sendMenu(
                "Hmmm, have you made up your mind? Choose the 2nd job advancement of your liking...",
                "Hunter",
                "Crossbow Man"
            )
            var jobName = (branch === 0) ? "Hunter" : "Crossbow Man"
            var jobId   = (branch === 0) ? 310      : 320

            if (npc.sendYesNo("So you want to advance as a #b" + jobName + "#k? Once you decide, you can't go back. Are you sure?")) {
                plr.setJob(jobId)
                if (branch === 0) {
                    npc.sendBackNext("Alright, you're a #bHunter#k from here on out. Train each day and hone your precision.", true, true)
                } else {
                    npc.sendBackNext("Alright! You have now become the #bCrossbow Man#k! Know your enemy's weakness, and strike true.", true, true)
                }
                npc.sendBackNext("Your #bUse#k and #bETC#k inventories have been expanded. Your Max HP and MP have also increased. Go check it out!", true, true)
                npc.sendOk("I have also given you a little bit of #bSP#k. Open the #bSkill Menu#k to enhance your 2nd job skills. Some skills require others first; remember that.")
                picked = branch
            } else {
                npc.sendOk("Take your time. This decision is important.")
            }
        }
    } else {
        npc.sendOk("Keep training as a Bowman. Return to me at #rLevel 30#k for your next advancement.")
    }

} else if (plr.job() === 310 || plr.job() === 320) {
    npc.sendOk("Walk the path you've chosen with pride. Keep training and growing stronger.")

} else {
    npc.sendOk("Those who want to become a bowman...")
}