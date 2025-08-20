if (plr.itemQuantity(5220000) < 1 && plr.itemQuantity(5451000) < 1) {
    npc.sendOk("You don't have a single ticket with you. Please buy the ticket at the department store before coming back to me. Thank you.")
} else {
    if (npc.sendYesNo("You have some #bGachapon Tickets#k there. \r\nWould you like to try your luck?")) {
        var prize = [2000005, 2022113, 2002018, 1382001, 1050008, 1442017, 1002084, 1050003, 1002064, 1061006, 1051027, 1442009, 1050056, 1051047, 1050049, 1040080, 1051055, 1372010, 1422005, 1002143, 1302027, 1061087, 1372003, 1302019, 1051023, 1050054, 1061083, 1051017, 1002028, 1322010, 1332013, 1050055, 1002245]
        var itemId = prize[Math.floor(Math.random() * prize.length)]
        var ticketId = (plr.itemQuantity(5220000) > 0 && plr.mapId() == 101000000) ? 5220000 : 5451000
        
        if (plr.takeItem(ticketId, 1)) {
            plr.giveItem(itemId, 1)
            npc.sendOk("You have obtained #b#t" + itemId + "##k.")
        } else {
            npc.sendOk("Please check your item inventory and see if you have the ticket, or if the inventory is full.")
        }
    }
}

// Generate by kimi-k2-instruct