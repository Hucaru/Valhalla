// Soldier Anderson – 922010000
var mapId = plr.getMapId()
var stage = Math.floor(mapId / 100) % 10

if (stage === 0) {
    // action0
    if (npc.sendYesNo("You'll have to start over from scratch if you want to take a crack at this quest after leaving this stage. Are you sure you want to leave this map?")) {
        plr.warp(922010000)
    } else {
        npc.sendBackNext("l see. Gather up the strength of your party members and try harder!", true, true)
    }
} else if (stage === 8) {
    // action2
    plr.warp(221023300)
    plr.takeItem(4001454, plr.itemQuantity(4001454))
} else {
    // action1
    var menu = "Do you need some help? \r\n#L0##bI need Platform Puppet.#l\r\n#L1#I want to get out of here.#l"
    npc.sendSelection(menu)
    var sel = npc.selection()

    if (sel === 0) {
        if (plr.itemQuantity(4001454) >= 1 || plr.canGainItem(4001454, 1)) {
            plr.giveItem(4001454, 1)
            npc.sendBackNext("You have received a Platform Puppet. If you place it on the platform, it will have the same effect as someone standing there. Remember, though, this is an item that can only be used in here.", true, true)
        } else {
            npc.sendBackNext("I can't give you the #bPlatform Puppet#k, because you don't have any room in your Inventory. Please empty 1 slot in your #rEtc#k window.", true, true)
        }
    } else if (sel === 1) {
        plr.warp(910340000)
    }
}

// Generate by kimi-k2-instruct