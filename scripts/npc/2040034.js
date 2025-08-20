npc.sendSelection("#e<Party Quest: Dimensional Crack>#n \r\nYou can't go any higher because of the extremely dangerous creatures above. Would you like to collaborate with party members to complete the quest? If so, please have your #bparty leader#k talk to me. \r\n#L0##bI want to participate in the party quest.#l\r\n#L1#I want to find party members.#l\r\n#L2#I want to receive the Broken Glasses.#l\r\n#L3#I would like to hear more details.#l")
var sel = npc.selection()

if (sel == 0) {
    if (plr.party() == null || !plr.isPartyLeader()) {
        npc.sendOk("From here on above, this place is full of dangerous objects and monsters, so I can't let you make your way up anymore. If you're interested in saving us and bring peace back into Ludibrium, however, that's a different story. If you want to defeat a powerful creature residing at the very top, then please gather up your party members. It won't be easy, but ... I think you can do it.")
    } else {
        var members = plr.partyMembers()
        if (members.length < 3) {
            npc.sendNext("You cannot participate in the quest because you do not have at least 3 party members.")
        } else {
            var allHere = true
            var allLevel = true
            for (var i = 0; i < members.length; i++) {
                if (members[i].mapId() != 221023300) allHere = false
                if (members[i].level() < 30) allLevel = false
            }
            if (!allHere) {
                npc.sendNext("Some of your party members are in a different map. Make sure they're all here!")
            } else if (!allLevel) {
                npc.sendNext("Either you or one of your party members is below Lv. 30. Please fit the level requirement before proceeding.")
            } else {
                var em = npc.getEventManager("LudiPQ")
                var prop = em.getProperty("state")
                if (prop == null || prop == 0) {
                    em.startInstance(plr.party(), plr.map(), 200)
                } else {
                    npc.sendNext("Another party is already fighting on the other side. Wait a moment and try again.")
                }
            }
        }
    }
} else if (sel == 1) {
    npc.sendPartyWindow()
} else if (sel == 2) {
    var quest = plr.getQuest(1202)
    var count = parseInt(quest.customData() || "0")
    if (count < 35) {
        npc.sendNext("I am offering 1 #v1022073##b#t1022073##k for every 35 times you help me. If you help me #b" + (35 - count) + " more times, you can receive Broken Glasses.")
    } else {
        if (npc.sendYesNo("Thank you for your help. You have helped " + count + " times in total and have received 0 of #bBroken Glasses(s)#k, so you still have 1 remaining. Would you like to receive #bBroken Glasses#k?")) {
            if (plr.getFreeSlots(1) < 1) {
                npc.sendOk("Equip item inventory is full.")
            } else {
                plr.giveItem(1022073, 1)
                quest.setCustomData(count - 35)
            }
        }
    }
} else if (sel == 3) {
    npc.sendOk("#e<Party Quest: Dimensional Crack>#n \r\nA Dimensional Crack has appeared in #bLudibrium#k! We desperately need brave adventurers who can defeat the intruding monsters. Please, party with some dependable allies to save Ludibrium! You must pass through various stages by defeating monsters and solving quizzes, and ultimately defeat #rAlishar#k. \r\n#e- Level#n: 30 or higher #r(Recommended Level: 20 - 69 )#k  \r\n#e- Time Limit#n: 20 min. \r\n#e- Number of Participants#n: 3 to 6 \r\n#e- Reward#n: #v1022073# #t1022073# #b(obtained every 35 time(s) you participate)#k \r\n                         Various Use, Etc, and Equip items")
}

// Generate by kimi-k2-instruct