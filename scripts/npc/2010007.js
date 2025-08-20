npc.sendBackNext("Hey...would you happen to be interested in GUILDS by any chance?", false, true)

var guildRank = plr.guildRank()
if (guildRank == 1) {
    // Guild leader menu
    var leaderMenu = "What do you want? Tell me... \r\n#L5##bI want to expand the guild.#l\r\n#L6#I want to disband the guild.#l\r\n#L7#I want to change the guild leader.#l"
    npc.sendSelection(leaderMenu)
    var leaderSel = npc.selection()

    if (leaderSel == 5) {
        npc.sendBackNext("Are you here because you want to expand your guild? To increase the number of people you can accept into your guild, you'll have to re-register. You'll also have to pay a fee. Just so you know, the absolute maximum size of a guild is 200 members.", true, true)
        if (npc.sendYesNo("Current Max Guild Members: 20 characters. To increase that amount by #b10#k, you need #b10000 GP#k. Your guild has #b150 GP#k right now. Do you want to expand your guild?")) {
            // expand guild capacity
            npc.sendOk("Guild capacity increased.")
        }
    } else if (leaderSel == 6) {
        if (npc.sendYesNo("Are you sure you want to break up your guild? Remember, once you break up your guild, it will be gone forever. Are you sure you still want to do it?")) {
            if (npc.sendYesNo("I'll ask one more time. Would you like to give up all guild privileges and disband the guild?")) {
                // disband guild
                npc.sendOk("Guild disbanded.")
            }
        }
    } else if (leaderSel == 7) {
        npc.sendBackNext("Is leading a guild becoming a burden on you? Select a new leader from the member list to appoint by right-clicking and pressing the Make GM button. However, the member you select must be online.", true, true)
    }
} else {
    // Non-guild leader menu
    var menu = "What do you want? Tell me... \r\n#L1##bWhat is a guild?#l\r\n#L2#How do I create a guild?#l\r\n#L3#I want to create a guild.#l"
    npc.sendSelection(menu)
    var sel = npc.selection()

    if (sel == 1) {
        npc.sendBackNext("You can think of a guild as a small crew full of people with similar interests and goals, except it will be officially registered in our Guild Base and be accepted as a valid GUILD.", false, true)
        npc.sendBackNext("There are a variety of benefits that you can get through guild activities. For example, you can obtain a guild skill or an item that is exclusive to guilds.", true, true)
    } else if (sel == 2) {
        npc.sendBackNext("You must be at least Lv. 101 to create a guild.", false, true)
        npc.sendBackNext("You also need 5,000,000 mesos. This is the registration fee.", true, true)
        npc.sendBackNext("So, come see me if you would like to register a guild! Oh, and of course you can't be already registered to another guild!", true, true)
    } else if (sel == 3) {
        if (npc.sendYesNo("Oh! So you're here to register a guild... You need 5,000,000 mesos to register a guild. I trust that you are ready. Would you like to create a guild?")) {
            if (plr.guildId() > 0) {
                npc.sendOk("You may not create a new Guild while you are in one.")
            } else if (plr.level() < 101) {
                npc.sendOk("Hey, your level is a bit low to be a guild leader. You need to be at least level 101 to create a guild.")
            } else if (plr.mesos() < 5000000) {
                npc.sendBackNext("Please check again. You'll need to pay the service fee to create and register a guild.", true, true)
            } else {
                var guildName = npc.askText("Enter the name of your guild, and your guild will be created. The guild will also be officially registered under our Guild Base, so best of luck to you and your guild!")
                // create guild
                npc.sendOk("Guild '" + guildName + "' has been created!")
            }
        }
    }
}

// Generate by kimi-k2-in