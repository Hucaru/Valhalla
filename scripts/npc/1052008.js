// Treasure Box (stateless, nested gating)

var pos = plr.position();

if (!pos || pos.x < -50 || pos.x > 250 || pos.y > 600) {
    npc.sendOk("You cannot see very well because you're too far. Go a little closer.");
} else if (!plr.giveItem(4031039, 1)) {
    npc.sendOk("Looking carefully into #p1052008#, there seems to be a shiny object inside but since your etc. inventory is full, that item is unattainable.");
} else {
    npc.sendBackNext(
        "Looking carefully into #p1052008#, there seems to be a shiny object inside. Reached out with a hand and was able to attain a small coin.",
        false,
        true
    );
    plr.warp(103020000, 0);
}
