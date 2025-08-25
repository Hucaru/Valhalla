npc.sendSelection("You can take the boat, if you like. Where would you like to go? \r\n#L0#Florina Beach, please.#l")
var sel0 = npc.selection()

npc.sendSelection("Have you heard about #b#m120030000##k not too far from #m120020400#? You can go there if you have #b2000 Mesos#k or a #bVIP Ticket to Florina Beach#k. \r\n#L0##bI'll pay 2000 Mesos.#l\r\n#L1#I have a VIP Ticket to Florina Beach.#l\r\n#L2#What is a VIP Ticket to Florina Beach?#l")
var sel = npc.selection()

if (sel == 0) {
    if (npc.sendYesNo("You want to pay #b2000 Mesos#k to go to #m120030000#? Sure, but don't forget that there are monsters there, too. I'll prepare to set sail. Okay. You ready to head for #m120030000# right now?")) {
        if (plr.mesos() < 2000) {
            npc.sendOk("I think you're lacking mesos. There are many ways to gather up some money, you know, like... selling your armor... defeating monsters... doing quests... you know what I'm talking about.")
        } else {
            plr.takeMesos(2000)
            plr.warp(120030000)
        }
    } else {
        npc.sendBackNext("You must have some business to take care of here. You must be tired from all that traveling and hunting. Go take some rest, and if you feel like changing your mind, then come talk to me.", true, true)
    }
} else if (sel == 1) {
    if (npc.sendYesNo("So you have a #bVIP Ticket to Florina Beach#k? You can always head over to Florina Beach with that. Alright then, but just be aware that you may be running into some monsters there too. Okay, would you like to head over to Florina Beach right now?")) {
        if (!plr.takeItem(4031134, 1)) {
            npc.sendOk("Hmmm, so where exactly is your #bVIP Ticket to Florina Beach#k? Are you sure you have one? Please double-check.")
        } else {
            plr.warp(120030000)
        }
    } else {
        npc.sendBackNext("You must have some business to take care of here. You must be tired from all that traveling and hunting. Go take some rest, and if you feel like changing your mind, then come talk to me.", true, true)
    }
} else if (sel == 2) {
    npc.sendBackNext("You must be curious about a #bVIP Ticket to Florina Beach#k. Haha, that's very understandable. A VIP Ticket to Florina Beach is an item where as long as you have in possession, you may make your way to Florina Beach for free. It's such a rare item that even we had to buy those, but unfortunately I lost mine a few weeks ago during my precious summer break.", false, true)
    npc.sendBackNext("I came back without it, and it just feels awful not having it. Hopefully someone picked it up and put it somewhere safe. Anyway, this is my story and who knows, you may be able to pick it up and put it to good use. If you have any questions, feel free to ask.", true, true)
}

// Generate by kimi-k2-instruct