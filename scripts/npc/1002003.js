npc.sendYesNo("I'm hoping for lots of business today! Friendly business from people looking to expand their Buddy List! You look like you might be... mildly popular. Give me some mesos and you can have even MORE friends? Just you though, not anybody else on your account.")
if (npc.sendYesNo("You're lucky! I'm giving you a #rmassive discount#k. #bIt'll be 50000 Mesos to permanently add 5 slots to your Buddy List#k. That's a deal for somebody as popular as you are! What do you say?")) {
    var capacity = plr.buddyCapacity()
    if (plr.mesos() < 50000 || capacity > 199) {
        npc.sendBackNext("Uh, you sure you got the money? It's #b50000 Mesos#k. Or maybe your Buddy List has already been fully expanded? No amount of money will make that list longer than #b200#k.", true, true)
    } else {
        plr.takeMesos(50000)
        plr.setBuddyCapacity(capacity + 5)
        npc.sendOk("You just got room for five more friends. I'll be here if you need to add more, but I'm not giving these things out for free.")
    }
} else {
    npc.sendBackNext("Are you broke, or just lonely?", true, true)
}

// Generate by kimi-k2-instruct