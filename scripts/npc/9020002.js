// Determine which action branch to use
var mapId = plr.mapId()
var reactor = (mapId % 1000 == 0) ? 2 : (mapId % 1000 == 300) ? 1 : 0

if (reactor == 0) {
    // action0 branch
    if (npc.sendYesNo("You'll have to start over from scratch if you want to take a crack at this quest after leaving this stage. Are you sure you want to leave this map?")) {
        plr.warp(910340000, 0)
    } else {
        npc.sendBackNext("l see. Gather up the strength of your party members and try harder!", true, true)
    }
} else if (reactor == 1) {
    // action1 branch
    var menu = "Do you need some help? \r\n#L0##bI need Platform Puppet.#l\r\n#L1#I want to get out of here.#l"
    npc.sendSelection(menu)
    var selection = npc.selection()

    if (selection == 0) {
        if (plr.getFreeSlots(4) < 1) {
            npc.sendBackNext("I can't give you the #bPlatform Puppet#k, because you don't have any room in your Inventory. Please empty 1 slot in your #rEtc#k window.", true, true)
        } else {
            plr.giveItem(4001454, 1)
            npc.sendBackNext("You have received a Platform Puppet. If you place it on the platform, it will have the same effect as someone standing there. Remember, though, this is an item that can only be used in here.", true, true)
        }
    } else if (selection == 1) {
        plr.warp(910340000, 0)
    }
} else if (reactor == 2) {
    // action2 branch
    plr.warp(910340700, 0)
    plr.takeItem(4001007, plr.itemQuantity(4001007))
}

// Generate by kimi-k2-instruct