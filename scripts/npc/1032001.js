/* 
** NPC: Grendel the Really Old
** Location: Ellinia — Magic Library
*/

if (plr.job() === 0) {
    if (plr.getLevel() >= 8) {
        npc.sendBackNext(
            "Do you want to be a Magician? It is an important and final choice. You will not be able to turn back.",
            false, true
        )
        if (npc.sendYesNo("You definitely have the look of a Magician. You may not be there yet, but I can see the Magician in you. Do you want to become a #rMagician#k?")) {
            plr.setJob(200)
            plr.giveItem(1372000, 1)
            npc.sendBackNext("Alright, you're a Magician from here on out, since I, Grendel the Really Old, allow you so. It isn't much, but I'll give you a little bit of what I have...", true, true)
            npc.sendBackNext("You have just equipped yourself with much more magical power. Please keep training and make yourself much better... I'll be watching you from here and there...", true, true)
            npc.sendOk("I also gave you a little bit of #bSP#k. Open the #bSkill Menu#k at the bottom-left to learn skills. Some skills require others first, so choose wisely.")
        } else {
            npc.sendOk("Come back once you have thought about it some more.")
        }
    } else {
        npc.sendOk("You need to train more. Return to me at #bLevel 8#k, and I will teach you the way of the #rMagician#k.")
    }

} else if (plr.job() === 200) {
    if (plr.getLevel() >= 30) {
        npc.sendBackNext("You got back here safely. Well done. I knew you'd pass your early trials very easily... now, I can make you much stronger.", false, true)

        var explain = npc.sendMenu(
            "When you are ready, choose your path.",
            "Please explain the role of the Wizard of Fire and Poison.",
            "Please explain the role of the Wizard of Ice and Lightning.",
            "Please explain the role of the Cleric.",
            "I'll choose my occupation!"
        )
        if (explain === 0) {
            npc.sendOk("Fire/Poison Wizards harness the destructive force of flame and the withering touch of toxins to decimate foes.")
        } else if (explain === 1) {
            npc.sendOk("Ice/Lightning Wizards command freezing blizzards and crackling lightning to control the battlefield.")
        } else if (explain === 2) {
            npc.sendOk("Clerics support allies with restorative magic and holy power, while smiting the undead.")
        }

        var picked = -1
        while (picked === -1) {
            var branch = npc.sendMenu(
                "Now, have you made up your mind? Choose the 2nd job advancement of your liking...",
                "The Wizard of Fire and Poison",
                "The Wizard of Ice and Lightning",
                "Cleric"
            )
            var jobName = (branch === 0) ? "Wizard of Fire and Poison"
                : (branch === 1) ? "Wizard of Ice and Lightning"
                    : "Cleric"
            var jobId   = (branch === 0) ? 210
                : (branch === 1) ? 220
                    : 230

            if (npc.sendYesNo("So you want to advance as the #b" + jobName + "#k? Once you decide, you can't go back. Are you sure?")) {
                plr.setJob(jobId)
                npc.sendBackNext("From here on out, you have become the #b" + jobName + "#k. Continue your studies, and I may one day make you even more powerful.", true, true)
                npc.sendBackNext("Your #bUse#k and #bETC#k inventories have been expanded. Your Max MP has also increased. Go check it out!", true, true)
                npc.sendOk("I have also given you a little bit of #bSP#k. Open the #bSkill Menu#k to enhance your 2nd job skills. Some skills require others first; remember that.")
                picked = branch
            } else {
                npc.sendOk("Take your time. This decision is important.")
            }
        }
    } else {
        npc.sendOk("Keep training as a Magician. Return to me at #rLevel 30#k for your next advancement.")
    }

} else if (plr.job() === 210 || plr.job() === 220 || plr.job() === 230) {
    npc.sendOk("Walk the path you've chosen with wisdom. Keep training and growing stronger.")
} else {
    npc.sendOk("To all that desire to become a magician... talk to me...")
}