// Check quest status first
if (plr.getQuestStatus(3239) != 1) {
    npc.sendOk("Lately the mechanical parts have been missing at the Toy Factory, and that really concerns me. I want to ask for help, but you don't seem strong enough to help us out. Who should I ask...?")
    return
}

// Initial briefing
npc.sendBackNext("Okay, then. Inside this room, you'll see a whole lot of plastic barrels lying around. Strike the barrels to knock them down, and see if you can find the lost #bMachine Parts#k inside. You'll need to collect 10 #bMachine Parts#k and then talk to me afterwards. There's a time limit on this, so go!", false, true)

// Check if instance is occupied
if (plr.getMap(922000000).getCharacters().size() >= 1) {
    npc.sendBackNext("I'm sorry, but it seems like someone else is inside looking through the barrels. Only one person is allowed in here, so you'll have to wait for your turn.", true, true)
    return
}

// Reset map and enter
plr.getMap(922000000).resetFully()
plr.warp(922000000)
plr.startMapTimeLimitTask(1200, 220020000)

// Later, when player returns with items
if (plr.itemQuantity(4031092) < 10) {
    var str = "Have you taken care of everything? If you wish to leave, I'll let you out. Ready to go? \r\n\r\n#L0##bPlease let me out."
    npc.sendSelection(str)
    var sel = npc.selection()
    
    if (npc.sendYesNo("Hmm... All right. I can let you out, but you'll have to start from the beginning next time. Still wanna leave?")) {
        plr.warp(922000009)
    }
} else {
    npc.sendBackNext("Oh ho, you really brought 10 Machine Parts items, and just in time. All right then! Since you have done so much for the toy factory, l'll give you a great present. Before I do that, however, make sure you have at least one empty slot in your Use tab.", false, true)
    
    if (plr.getInventoryFreeSlots(2) < 1) {
        npc.sendOk("Use item inventory is full.")
        return
    }
    
    plr.gainExp(140874)
    plr.giveItem(2040708, 1)
    plr.takeItem(4031092, 10)
    plr.warp(220020000)
}

// Generate by kimi-k2-instruct