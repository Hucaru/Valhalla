var eim = plr.instanceProperties();
var isLeader = plr.isPartyLeader();
var stage4 = +eim.stage4 || 0;

if (isLeader && stage4) {
    // Reactor = action1 flow
    npc.sendNext("Wow, not a single #b#o9300010##k left! I'm impressed! I can open the portal to the next stage now.")
    npc.sendBackNext("The portal that leads you to the next stage is now open.", true, true)
    eim.stage4 = "1";
    // Portal open visual handled engine-side
} else {
    // Reactor = action0 flow
    npc.sendNext("Welcome to the fourth stage. Here, you must face the powerful #b#o9300010##k. #b#o9300010##k is a fearsome opponent, so do not let your guard down. Once you defeat it, let me know and I'll show you to the next stage.")
}