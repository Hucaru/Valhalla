// Determine which map the player is on
var mapId = plr.mapId()

if (mapId == 103000000) {
    // First map branch
    npc.sendSelection("#e<Party Quest: First Time Together>#n \r\nHow would you like to complete a quest by working with your party members? Inside, you will find many obstacles that you will have to overcome with the help of your party members. \r\n#L0##bGo to the First Accompaniment Lobby.#l")
    var sel = npc.selection()
    if (sel == 0) {
        plr.saveLocation("MULUNG_TC")
        plr.warp(910340700)
    }
} else {
    // Second map branch
    npc.sendSelection("#e<Party Quest: First Time Together>#n \r\nInside, you'll find many obstacles that can only be solved by working with a party. Interested? Then have your #bParty Leader#k talk to me. \r\n#L2##bI want to do the Party Quest.#l\r\n#L3#I want to find party members who will join me.#l\r\n#L4#I want to hear the details.#l\r\n#L5#I want a Soft Jelly Shoes.#l")
    var sel = npc.selection()

    if (sel == 2) {
        // Check party leader
        if (!plr.isPartyLeader()) {
            npc.sendOk("You can't enter alone. If you want to enter, the party leader must talk to me. Ask your party leader to do so.")
            return
        }

        // Check party size
        var party = plr.getPartyMembers()
        if (party.length < 3) {
            npc.sendBackNext("You cannot enter because your party doesn't have 3 members. You need 3 party members at Lv. 20 or higher to enter, so double-check and talk to me again.", true, true)
            return
        }

        // Check all members on same map
        for (var i = 0; i < party.length; i++) {
            if (party[i].mapId != 910340700) {
                npc.sendBackNext("Some of your party members are in a different map. Make sure they're all here!", true, true)
                return
            }
        }

        // Check all members level 20+
        for (var i = 0; i < party.length; i++) {
            if (party[i].level < 20) {
                npc.sendBackNext("Some members of your party haven't reached Lv. 20 yet. Your party must have 3 players who are at least Lv. 20 characters to enter the area. Talk to me again when your party is ready.", true, true)
                return
            }
        }

        // Start PQ
        var started = plr.startPartyEvent("KerningPQ")
        if (!started) {
            npc.sendBackNext("Some other party has already gotten in to try clearing the quest. Please try again later.", true, true)
        }

    } else if (sel == 3) {
        plr.openPartyWindow()
    } else if (sel == 4) {
        npc.sendOk("Calling on all those with courage! Work together, share your strengths, and use your wisdom to find and defeat the vicious #rKing Slime#k! King Slime will appear once you complete certain challenges, such as collecting passes or answering location based quizzes. \r\n#e- Level#n: 20 or higher #r(Recommended Level: 20 - 69 )#k  \r\n#e- Time Limit#n: 20 min. \r\n#e- Number of Participants#k: 3 to 4 \r\n#e- Reward#n: #v1072369# Squishy Shoes #b(dropped by King Slime)#k  \r\n#e- Reward#n: #v1072533# Soft Jelly Shoes #b(exchanged for Smooshy Liquid x15)#k")
    } else if (sel == 5) {
        if (plr.itemQuantity(4001531) < 15) {
            npc.sendOk("If you want #v1072533# Soft Jelly Shoes, you'll need 15 #b#t4001531#s#k. You can obtain Smooshy Liquids by defeating #rKing Slime.")
            return
        }
        if (plr.getEquipInventoryFreeSlots() < 1) {
            npc.sendOk("Equip item inventory is full.")
            return
        }
        plr.takeItem(4001531, 15)
        plr.giveItem(1072533, 1)
        npc