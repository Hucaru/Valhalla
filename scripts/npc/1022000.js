/* 
** NPC: Dances With Balrog
** Location: Perion: Warriors' Sanctuary
*/

// Beginner -> Warrior (Level 10+)
if (plr.job() === 0) {
    if (plr.getLevel() >= 10) {
        npc.sendBackNext(
            "Do you wish to become a Warrior? It is an important and final choice. You will not be able to turn back.",
            false, true
        )
        if (npc.sendYesNo("You definitely have the look of a Warrior. You may not be there yet, but I can see the Warrior in you. Do you want to become a #rWarrior#k?")) {
            plr.setJob(100)
            // Starter sword
            plr.giveItem(1302000, 1)
            npc.sendBackNext("Alright! You are a Warrior from here on out... Here's a little bit of my power to you... Haahhhh!", true, true)
            npc.sendBackNext("I have added slots for your equipment and etc. inventory. You have also gotten much stronger. Train harder, and you may one day reach the very top. I'll be watching you from afar. Please work hard.", true, true)
            npc.sendOk("I also gave you a little bit of #bSP#k. Open the #bSkill Menu#k at the bottom-left to learn skills. Some skills require others first, so choose wisely.")
        } else {
            npc.sendOk("Come back once you have thought about it some more.")
        }
    } else {
        npc.sendOk("You need to train more. Return to me at #bLevel 10#k, and I will teach you the way of the #rWarrior#k.")
    }

// Warrior (1st job) -> 2nd job (Level 30+): Fighter / Page / Spearman
} else if (plr.job() === 100) {
    if (plr.getLevel() >= 30) {
        npc.sendBackNext("Whoa, you have definitely grown up! You don't look small and weak anymore... I can feel your presence as a Warrior!", false, true)

        var choice = npc.sendMenu(
            "When you are ready, choose your path.",
            "Please explain the role of the Fighter.",
            "Please explain the role of the Page.",
            "Please explain the role of the Spearman.",
            "I'll choose my occupation!"
        )
        if (choice === 0) {
            npc.sendOk("Fighters focus on raw strength and direct combat, pushing through enemies with overwhelming power.")
        } else if (choice === 1) {
            npc.sendOk("Pages employ tactical strikes and elemental prowess to exploit enemy weaknesses.")
        } else if (choice === 2) {
            npc.sendOk("Spearmen wield polearms or spears, striking from reach and bolstering themselves and allies.")
        }

        var branch = -1
        while (branch === -1) {
            var ready = npc.sendMenu(
                "Hmmm, have you made up your mind? Choose the 2nd job advancement of your liking...",
                "Fighter",
                "Page",
                "Spearman"
            )
            var jobName = (ready === 0) ? "Fighter" : (ready === 1) ? "Page" : "Spearman"
            var jobId   = (ready === 0) ? 110      : (ready === 1) ? 120    : 130

            if (npc.sendYesNo("So you want to advance as a #b" + jobName + "#k? Once you decide, you can't go back. Are you sure?")) {
                plr.setJob(jobId)
                npc.sendBackNext("Alright! You are now a #b" + jobName + "#k! Keep training and hone your skills.", true, true)
                npc.sendOk("I have also given you a little bit of #bSP#k. Open the #bSkill Menu#k to enhance your 2nd job skills. Some skills require others first; remember that.")
                branch = ready
            } else {
                npc.sendOk("Take your time. This decision is important.")
            }
        }
    } else {
        npc.sendOk("Keep training as a Warrior. Return to me at #rLevel 30#k for your next advancement.")
    }

// Already 2nd job Warrior
} else if (plr.job() === 110 || plr.job() === 120 || plr.job() === 130) {
    npc.sendOk("Walk the path you've chosen with pride. Keep training and growing stronger.")

// Other classes
} else {
    npc.sendOk("For those that want to become a Warrior, come see me...")
}