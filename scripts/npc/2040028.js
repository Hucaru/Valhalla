npc.sendSelection("Hello, there. I'm #b#p2040028##k, in charge of protecting this room. Inside, you'll see a bunch of dollhouses, and you may find one that looks a little bit different from the others. Your job is to locate it, break its door, and find the #bPendulum#k, which is an integral part of the Ludibrium Clocktower. You'll have a time limit on this, and if you break the wrong dollhouse, you'll be forced back outside, so please be careful. \r\n#L0##bI want to get out of here.#l")
var sel = npc.selection()

if (sel == 0) {
    if (npc.sendYesNo("Are you sure you want to give up now? Alright then... but please remember that the next time you visit this place, the dollhouses will switch places, and you'll have to look through each and every one of them carefully again. What do you think? Would you still like to leave this place?")) {
        plr.warp(221023200)
    } else {
        npc.sendBackNext("I knew you'd stay. It's important that you finish what you've started! Now please go locate the different-looking dollhouse, break it, and bring #bPendulum#k to me!", true, true)
    }
}

// Generate by kimi-k2-instruct