// Warrior Job Instructor
// Entry check
var job_ok   = plr.job() === 100
var lvl_ok   = plr.getLevel() >= 30
var letter   = plr.itemCount(4031008) >= 1
var darkCnt  = plr.itemCount(4031013)
var proof    = plr.itemCount(4031012) === 0

var qualified = job_ok && lvl_ok && letter && (darkCnt < 30) && proof
if (!qualified) {
    npc.sendOk("What?... Do you want something from me?...")
}

// Show intro sequence
npc.sendBackNext(
    "Hmmm...it is definitely the letter from #bDances with Balrog#k...so you came all the way here to take the test and make the 2nd job advancement as the warrior. Alright, I'll explain the test to you. Don't sweat it too much, it's not that complicated.",
    false, true
)
npc.sendBackNext(
    "I'll send you to a hidden map. You'll see monsters you don't normally see. They look the same like the regular ones, but with a totally different attitude. They neither boost your experience level nor provide you with item.",
    true, true
)
npc.sendBackNext(
    "You'll be able to acquire a marble called #bDark Marble#k while knocking down those monsters. It is a special marble made out of their sinister, evil minds. Collect 30 of those, and then go talk to a colleague of mine in there. That's how you pass the test.",
    true, true
)

var go = npc.sendYesNo(
    "Once you go inside, you can't leave until you take care of your mission. If you die, your experience level will decrease..so you better really buckle up and get ready...well, do you want to go for it now?"
)

if (!go) {
    npc.sendOk("Really? Have to give more though to it, huh? Take your time, take your time. This is not something you should take lightly...come talk to me once you have made your decision.")
}

// Choose the first empty chamber
if (cm.getPlayerCount(108000300) === 0) {
    plr.warp(108000300)
} else if (cm.getPlayerCount(108000301) === 0) {
    plr.warp(108000301)
} else if (cm.getPlayerCount(108000302) === 0) {
    plr.warp(108000302)
} else {
    npc.sendOk("I am sorry, but all test chambers are currently taken, you will have to wait until the people inside them finish their job advancement.")
}

npc.sendOk("Defeat the monsters inside, collect 30 Dark Marbles, then strike up a conversation with a colleague of mine inside. He'll give you #bThe Proof of a Hero#k, the proof that you've passed the test. Best of luck to you.")