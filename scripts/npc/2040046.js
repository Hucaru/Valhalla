npc.sendYesNo("I hope I can make as much as yesterday ...well, hello! Don't you want to extend your buddy list? You look like someone who'd have a whole lot of friends... well, what do you think? With some money I can make it happen for you. Remember, though, it only applies to one character at a time, so it won't affect any of your other characters on your account. Do you want to do it?")
if (npc.sendYesNo("Alright, good call! It's not that expensive actually. For a special Ludibrium discount, #b50000 mesos and I'll add 5 more slots to your buddy list#k. And no, I won't be selling them individually. Once you buy it, it's going to be permanently on your buddy list. So if you're one of those that needs more space there, then you might as well do it. What do you think? Will you spend #b50000 mesos#k for it?")) {
    var capacity = plr.buddyCapacity()
    if (plr.mesos() < 50000 || capacity > 199) {
        npc.sendOk("Hey... are you sure you have #b50000 mesos#k?? If so then check and see if you have extended your buddy list to the max. Even if you pay up, the most you can have on your buddy list is #b200#k.")
    } else {
        plr.takeMesos(50000)
        plr.setBuddyCapacity(capacity + 5)
        npc.sendOk("Alright! Your buddy list will have 5 extra slots by now. Check and see for it yourself. And if you still need more room on your buddy list, you know who to find. Of course, it isn't going to be for free ... well, so long ...")
    }
} else {
    npc.sendBackNext("I see... I don't think you don't have as many friends as I thought you would. If not, you just don't have #b50000 mesos with you right this minute? Anyway, if you ever change your mind, come back and we'll talk business. That is, of course, once you have get some financial relief ... hehe ...", true, true)
}

// Generate by kimi-k2-instruct