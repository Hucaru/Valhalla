// Decide which flow we are in (top-level selection already done by reactor string, script always runs once)
var mapId = plr.getMapId();

// ACTION2 : map % 1000 == 0   → map ends with 000 (ticket stage?)
if (mapId % 1000 == 0) {
    plr.warp(910340700);
    plr.removeItemsByID(4001007, plr.itemCount(4001007));
}

// ACTION1 : map % 1000 == 300 → 300 stage (help & puppet)
if (mapId % 1000 == 300) {
    var sel = npc.sendMenu(
        "Do you need some help?",
        "I need Platform Puppet.",
        "I want to get out of here."
    );

    if (sel === 0) {
        // Platform Puppet
        if (plr.itemCount(4001454) < 1) {
            if (plr.giveItem(4001454, 1)) {
                npc.sendNext("You have received a Platform Puppet. If you place it on the platform, it will have the same effect as someone standing there. Remember, though, this is an item that can only be used in here.");
            } else {
                npc.sendNext("I can't give you the #bPlatform Puppet#k, because you don't have any room in your Inventory. Please empty 1 slot in your #rEtc#k window.");
            }
        } else {
            npc.sendNext("You already have a Platform Puppet.");
        }
    } else {
        // Leave to lobby
        plr.warp(910340000);
    }
}

// ACTION0 : default (ends with anything else)
if (npc.sendYesNo("You'll have to start over from scratch if you want to take a crack at this quest after leaving this stage. Are you sure you want to leave this map?")) {
    plr.warp(910340000);
}