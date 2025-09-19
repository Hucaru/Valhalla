// Treasure Box – distance + inventory-safe flow (stateless)

var pos = plr.position();

// Too far to interact
if (!pos || pos.x < -50 || pos.x > 250 || pos.y > 600) {
    npc.sendOk("You cannot see very well because you're too far. Go a little closer.");
} else {
    // Try to give the item first; if it fails, inventory is full
    if (!plr.giveItem(4031040, 1)) {
        if (npc.sendNext("Looking carefully into Treasure Box, there seems to be a shiny object inside but since your etc. inventory is full, that item is unattainable.")) {
            if (typeof plr.mapID === "function" && plr.mapID() === 910360102) {
                plr.warp(103020000);
            }
        }
    } else {
        if (npc.sendNext("Looking carefully into Treasure Box, there seems to be a stack of papers in there. I reached out my hand and voila, a huge stack of money.")) {
            plr.warp(103020000);
        }
    }
}