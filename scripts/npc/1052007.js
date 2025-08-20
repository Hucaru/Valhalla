npc.sendSelection("Pick your destination. #b\r\n#L0#Subway Construction Site#l\r\n#L1#Kerning city Subway#r(Beware of Stirges and Wraiths!)#l\r\n#L2##bKerning Square Shopping Center (Get on the Subway).#l\r\n\r\n#L3#Enter Construction Site#l\r\n#L4#New Leaf City#l")
var select = npc.selection()

if (select == 0) {
    npc.sendOk("NPC ID is not yet implemented.")
} else if (select == 1) {
    plr.warp(103020100, 2)
} else if (select == 2) {
    plr.warp(103020010, 0)
    npc.sendOk("The next stop is at Kerning Square Station. The exit is to your left.")
} else if (select == 3) {
    if (!plr.hasItem(4031036) && !plr.hasItem(4031037) && !plr.hasItem(4031038)) {
        npc.sendOk("Here's the ticket reader. You are not allowed in without the ticket.")
    } else {
        var str = "Here's the ticket reader. You will be brought in immediately. Which ticket would you like to use? #b"
        if (plr.hasItem(4031036)) str += "\r\n#L0#Construction site B1#l"
        if (plr.hasItem(4031037)) str += "\r\n#L1#Construction site B2#l"
        if (plr.hasItem(4031038)) str += "\r\n#L2#Construction site B3#l"
        npc.sendSelection(str)
        var idx = npc.selection()
        plr.takeItem(item[idx], 1)
        plr.warp(map[idx], 0)
    }
} else if (select == 4) {
    if (!plr.hasItem(4031711)) {
        npc.sendOk("Here's the ticket reader. You are not allowed in without the ticket.")
    } else {
        npc.sendSelection("Here's the ticket reader. You will be brought in immediately. Which ticket would you like to use? \r\n#L0##bNew Leaf city (Normal)#l")
        var nlcSel = npc.selection()
        if (npc.sendYesNo("It looks like there's plenty of room for this ride. Please have your ticket ready so I can let you in. The ride will be long, but you'll get to your destination just fine. What do you think? Do you want to get on this ride?")) {
            plr.takeItem(4031711, 1)
            plr.warp(600010004, 0)
        }
    }
}

// Generate by kimi-k2-instruct