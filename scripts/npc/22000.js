// Shanks at Maple Port
if (plr.itemQuantity(4031801) > 0) {
    // Has recommendation letter
    if (npc.sendYesNo("Take this ship and you'll head off to a bigger continent. For #e150 mesos#n, I'll take you to #bVictoria Island#k. The thing is, once you leave this place, you will never be able to return. So, choice is yours. Do you want to go to Victoria Island?")) {
        npc.sendBackNext("Okay, now give me 150 mesos...Hey, what's that? Is that the recommendation letter from Lucas, the chief of Amherst? You should have told me about this earlier. I, Shanks, recognize greatness when l see it, and since you have been recommended by Lucas, l can see that you have very great potential as an adventurer. No way would l dare charge you for this trip!", false, true)
        npc.sendBackNext("Since you have the recommendation letter, I won't charge you for this. We're going to head to Victoria Island right now, so buckle up! it might get a bit turbulent!", true, true)
        plr.takeItem(4031801, 1)
        plr.warp(2010000)
    } else {
        npc.sendOk("Hmm... I guess you still have things to do here?")
    }
} else {
    // No recommendation letter
    if (npc.sendYesNo("You can go to a wider continent by getting on this ship. I will take you to #bVictoria Island#k for #e150 mesos#n. However, once you leave this place, you cannot come back. What do you think? Would you like to go to Victoria Island?")) {
        npc.sendBackNext("I bet you are bored of this place. Well... first, give me #e150 mesos#n.", false, true)
        if (plr.mesos() < 150) {
            npc.sendOk("What? You're telling me you wanted to go without any money? You're one weirdo...")
        } else {
            npc.sendBackNext("Awesome! #e150 mesos#n accepted! Alright, off to #bVictoria Island#k!", true, true)
            plr.takeMesos(150)
            plr.warp(2010000)
        }
    } else {
        npc.sendOk("Hmm... I guess you still have things to do here?")
    }
}

// Generate by kimi-k2-instruct