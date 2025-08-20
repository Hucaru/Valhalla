// Check distance
if (plr.getPosition().y > -1586) {
    npc.sendOk("You're too far from Trainer Frod. Get closer.")
    return
}

// Check for letter
if (plr.itemQuantity(4031035) < 1) {
    npc.sendOk("My brother told me to take care of the pet obstacle course, but ... since I'm so far away from him, I can't help but wanting to goof around ...hehe, since I don't see him in sight, might as well just chill for a few minutes.")
    return
}

// First dialogue
npc.sendBackNext("Eh, that's my brother's letter! Probably scolding me for thinking I'm not working and stuff...Eh? Ahhh...you followed my brother's advice and trained your pet and got up here, huh? Nice!! Since you worked hard to get here, I'll boost your intimacy level with your pet.", false, true)

// Check pet
if (plr.getPet(0) != null) {
    plr.takeItem(4031035, 1)
    var closenessValues = [1, 2, 3, 4, 5, 6, 7, 8, 9]
    var randomCloseness = closenessValues[Math.floor(Math.random() * closenessValues.length)]
    plr.gainClosenessAll(randomCloseness)
    npc.sendOk("What do you think? Don't you think you have gotten much closer with your pet? If you have time, train your pet again on this obstacle course...of course, with my brother's permission.")
} else {
    npc.sendBackNext("Hmmm ... did you really get here with your pet? These obstacles are for pets. What are you here for without it?? Get outta here!", true, true)
}

// Generate by kimi-k2-instruct