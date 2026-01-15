// Stage 6 NPC - LudiPQ (Jump Quest)
if (!plr.isPartyLeader()) {
    npc.sendOk("Here is the information about the 6th stage. Here, you'll see boxes with numbers written on them, and if you stand on top of the correct box by pressing the UP ARROW, you'll be transported to the next box. The party leader gets a clue #bonly twice#k. Once you reach the top, you'll find the portal to the next stage.");
} else {
    npc.sendOk("Navigate through the platforms to reach the portal to the next stage! The clue is: One, 3, 3, 2, middle, 1, three, 3, 3, left, two, 3, 1, one, ?");
}
