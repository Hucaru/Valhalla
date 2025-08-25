if (plr.level() < 20) {
    npc.sendOk("You can enter if you purchase a ticket, but it will be too much for you. Please come again after you make some more preparations. There's no telling what devices are set up deep underground!")
} else {
    var menu = "You must purchase the ticket to enter. Once you have made the purchase, you can enter through #p1052007# on the right. What would you like to buy? #b\r\n"
    menu += "#L0#Ticket to Construction Site B1#l\r\n"
    menu += "#L1#Ticket to Construction Site B2#l\r\n"
    if (plr.level() > 99) {
        menu += "#L2#Ticket to Construction Site B3#l"
    }
    npc.sendSelection(menu)
    var select = npc.selection()

    if (npc.sendYesNo("Will you purchase the Ticket to #bConstruction Site B" + (select + 1) + "#k? It'll cost you " + [500, 1200, 2000][select] + " Mesos. Before making the purchase, please make sure you have an empty slot on your ETC inventory.")) {
        if (plr.mesos() < [500, 1200, 2000][select] || !plr.giveItem([4031036, 4031037, 4031038][select], 1)) {
            npc.sendOk("Are you lacking Mesos? Check and see if you have an empty slot on your ETC inventory or not.")
        } else {
            plr.takeMesos([500, 1200, 2000][select])
            if (select == 0) {
                npc.sendOk("You can insert the ticket in the #p1052007#. I heard Area 1 has some precious items available but with so many traps all over the place most come back out early. Wishing you the best of luck.")
            } else if (select == 1) {
                npc.sendOk("You can insert the ticket in the #p1052007#. I heard Area 2 has rare, precious items available but with so many traps all over the place most come back out early. Please be safe.")
            } else if (select == 2) {
                npc.sendOk("You can insert the ticket in the #p1052007#. I heard Area 3 has very rare, very precious items available but with so many traps all over the place most come back out early. Wishing you the best of luck.")
            }
        }
    } else {
        npc.sendOk("You can enter the premise once you have bought the ticket. I heard there are strange devices in there everywhere but in the end, rare precious items await you. So let me know if you ever decide to change your mind.")
    }
}

// Generate by kimi-k2-instruct