if (plr.itemQuantity(5220000) < 1 && plr.itemQuantity(5451000) < 1) {
    npc.sendOk("You don't have a single ticket with you. Please buy the ticket at the department store before coming back to me. Thank you.")
} else {
    if (npc.sendYesNo("You have some #bGachapon Tickets#k there. \r\nWould you like to try your luck?")) {
        var prize = [2040317, 3010013, 2000005, 2022113, 2043201, 2044001, 2041038, 2041039, 2041036, 2041037, 2041040, 2041041, 2041026, 2041027, 2044600, 2043301, 2040308, 2040309, 2040304, 2040305, 2040810, 2040811, 2040812, 2040813, 2040814, 2040815, 2040008, 2040009, 2040010, 2040011, 2040012, 2040013, 2040510, 2040511, 2040508, 2040509, 2040518, 2040519, 2040520, 2040521, 2044401, 2040900, 2040902, 2040908, 2040909, 2044301, 2040406, 2040407, 1302026, 1061054, 1452003, 1382037, 1302063, 1041067, 1372008, 1432006, 1332053, 1432016, 1302021, 1002393, 1051009, 1082148, 1102082, 1061043, 1452005, 1051016, 1442012, 1372017, 1332000, 1050026, 1041062]
        var z = prize[Math.floor(Math.random() * prize.length)]
        var ticketId = (plr.itemQuantity(5220000) > 0 && plr.mapId() == 100000100) ? 5220000 : 5451000
        
        if (plr.takeItem(ticketId, 1)) {
            plr.giveItem(z, 1)
            npc.sendOk("You have obtained #b#t" + z + "##k.")
        } else {
            npc.sendOk("Please check your item inventory and see if you have the ticket, or if the inventory is full.")
        }
    }
}

// Generate by kimi-k2-instruct