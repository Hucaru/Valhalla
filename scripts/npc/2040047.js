// 士兵安德森 ‑ 遗弃之塔<冒险的终结> 送至922010000

var stage = (plr.getMap() / 100 % 10 == 0) ? 2 :
            (plr.getMap() / 100 % 10 == 8) ? 1 : 0;

if (stage === 0) {
    if (npc.sendYesNo("You'll have to start over from scratch if you want to take a crack at this quest after leaving this stage. Are you sure you want to leave this map?")) {
        plr.warp(922010000);
    }
} else if (stage === 1) {
    var choice = npc.sendMenu("Do you need some help?",
        "I need Platform Puppet.",
        "I want to get out of here."
    );

    if (choice === 0) {
        if (plr.itemCount(4001454) < 1) {
            npc.sendNext("I can't give you the #bPlatform Puppet#k, because you don't have any room in your Inventory. Please empty 1 slot in your #rEtc#k window.");
        } else {
            plr.giveItem(4001454, 1);
            npc.sendNext("You have received a Platform Puppet. If you place it on the platform, it will have the same effect as someone standing there. Remember, though, this is an item that can only be used in here.");
        }
    } else {
        plr.warp(910340000);
    }
} else if (stage === 2) {
    plr.warp(221023300);
    if (plr.itemCount(4001454) > 0) {
        plr.removeItemsByID(4001454, plr.itemCount(4001454));
    }
}