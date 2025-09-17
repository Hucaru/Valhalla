var eim = plr.instanceProperties();

if (eim.stage1 == null || !plr.isPartyLeader()) {
    npc.sendOk("In the first stage, you'll find Ratz and Black Ratz from Another Dimension, who are nibbling away at the Dimensional Schism. If you gather up 20 passes that the Ratz and Black Ratz have stolen, I'll open the way to the next stage. Good luck!");
} else {
    npc.sendOk("Wow! Congratulations on clearing the quests for this stage. Please use the portal you see over there and move on to the next stage. Best of luck to you!");
}