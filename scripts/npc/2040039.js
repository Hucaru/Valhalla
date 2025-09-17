var eim = plr.instanceProperties();

if (eim["stage2"] == null || !plr.isPartyLeader()) {
    npc.sendOk("In the second stage, the Dimensional Schism has spawned a place of pure darkness. Monsters called #b#o9300008##k have hidden themselves in the darkness. Defeat all of them, and then talk to me to proceed to the next stage.");
} else {
    npc.sendOk("Congratulations on clearing the quests for this stage. Please use the portal you see over there and move on to the next stage.");
}