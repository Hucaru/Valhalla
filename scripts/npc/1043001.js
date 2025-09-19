// Herb patch handler
// Herb patch handler (stateless/nested flow)

var pos = plr.position();

if (!pos || pos.y > -2962) {
    npc.sendOk("You can't see the inside of the pile of flowers very well because you're too far. Go a little closer.");
} else if (plr.itemCount(4031032) >= 1) {
    npc.sendOk("You already have #b#t4031032##k.");
} else if (npc.sendYesNo("Are you sure you want to take #b#t4031032##k with you?")) {
    if (plr.giveItem(4031032, 1)) {
        plr.warp(101000000);
    } else {
        npc.sendOk("Please make room in your Etc inventory.");
    }
}