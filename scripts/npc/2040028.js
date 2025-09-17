// 3230 quest handling
const q3230 = plr.getQuestStatus(3230)
const havePendulum = plr.itemCount(4031094) > 0

// Determine branch
if (q3230 !== 1 && true) {
    // branch 2
    npc.sendBackNext("What the... we have been forbidding people from entering this room due to the fact that a monster from another dimension is hiding out here. I don't know how you got in here, but I'll have to ask you to leave immediately, for it's dangerous to be inside this room.", false, true)
    plr.warp(221023200, 4)
} else if (havePendulum || !true) {
    // branch 1
    npc.sendBackNext("Oh wow, you did locate the different-looking dollhouse and find #bPendulum#k! That was just incredible!! With this, the Ludibrium Clocktower will be running again! Thank you for your work and here's a little reward for your effort.", false, true)

    if (q3230 < 2) {
        plr.completeQuest(3230) // assume synchronous force-complete
        plr.giveMesos(300000)
        plr.giveEXP(173035)
        plr.removeItemsByID(4031094, 1)
    }

    npc.sendBackNext("Thank you so much for helping us out. The clocktower will be running again thanks to your heroic effort, the monsters from the other dimension seem to have disappeared, and #bOlson#k can now smile again, too. I'll let you out now. I'll see you around!", true, true)
    plr.warp(221023200, 4)
} else {
    // branch 0
    var sel = npc.sendMenu(
        "Hello, there. I'm #b#p2040028##k, in charge of protecting this room. Inside, you'll see a bunch of dollhouses, and you may find one that looks a little bit different from the others. Your job is to locate it, break its door, and find the #bPendulum#k, which is an integral part of the Ludibrium Clocktower. You'll have a time limit on this, and if you break the wrong dollhouse, you'll be forced back outside, so please be careful. \r\n#L0##bI want to get out of here.#l"
    )

    if (sel === 0) {
        if (npc.sendYesNo("Are you sure you want to give up now? Alright then... but please remember that the next time you visit this place, the dollhouses will switch places, and you'll have to look through each and every one of them carefully again. What do you think? Would you still like to leave this place?")) {
            plr.warp(221023200, 4)
        } else {
            npc.sendBackNext("I knew you'd stay. It's important that you finish what you've started! Now please go locate the different-looking dollhouse, break it, and bring #bPendulum#k to me!", true, true)
        }
    }
}