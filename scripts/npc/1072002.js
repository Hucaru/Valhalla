// Bowman Job Instructor – stateless version
//
// Pre-check: Archer 300 / level ≥ 30
//            Must have 1 × 4031010 (Athena's Letter)
//            Must have < 30 × 4031013 (Dark Marble)
//            Must have 0 × 4031012 (Hero’s Proof)

if (!(plr.job() === 300 && plr.getLevel() >= 30 &&
      plr.itemCount(4031010) >= 1 &&
      plr.itemCount(4031013) < 30 &&
      plr.itemCount(4031012) === 0)) {

    npc.sendOk("Hello there. Do you want something from me?...")
} else {
    npc.sendBackNext("Hmmm...it is definitely the letter from #bAthena Pierce#k...so you came all the way here to take the test and make the 2nd job advancement as the Bowman. Alright, I'll explain the test to you. Don't sweat it too much, it's not that complicated.", false, true)
    npc.sendBackNext("I'll send you to a hidden map. You'll see monsters you don't normally see. They look the same like the regular ones, but with a totally different attitude. They neither boost your experience level nor provide you with item.", true, true)
    npc.sendBackNext("You'll be able to acquire a marble called #b#t4031013##k while knocking down those monsters. It is a special marble made out of their sinister, evil minds. Collect 30 of those, and then go talk to a colleague of mine in there. That's how you pass the test.", true, true)

    if (npc.sendYesNo("Once you go inside, you can't leave until you take care of your mission. If you die, your experience level will decrease..so you better really buckle up and get ready...well, do you want to go for it now?")) {
        // pick first free room among 108000100, 108000101, 108000102
        var room = 108000100
        if (plr.getPlayerCount(room) > 0) {
            room = 108000101
            if (plr.getPlayerCount(room) > 0) {
                room = 108000102
                if (plr.getPlayerCount(room) > 0) {
                    npc.sendOk("I am sorry, but all test chambers are currently taken, you will have to wait until the people inside them finish their job advancement.")
                }
            }
        }
        plr.warp(room)
        npc.sendOk("Defeat the monsters inside, collect 30 Dark Marbles, then strike up a conversation with a colleague of mine inside. He'll give you #bThe Proof of a Hero#k, the proof that you've passed the test. Best of luck to you.")
    } else {
        npc.sendBackNext("Really? Have to give more thought to it, huh? Take your time, take your time. This is not something you should take lightly...come talk to me once you have made your decision.", true, true)
    }
}