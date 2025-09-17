// 第7階段白花簇脚本
var etcSlot = 0;
var etcKey = 4; // MapleInventoryType.ETC

// Check distance
if (plr.getY() <= -3165) {
    etcSlot = plr.getFreeSlots(etcKey);
    if (etcSlot >= 1) {
        if (plr.giveItem(4031028, 30)) {
            plr.warp(105000000);
        }
    } else {
        npc.sendOk("Etc item inventory is full.");
    }
} else {
    npc.sendOk("You can't see the inside of the pile of flowers very well because you're too far. Go a little closer.");
}