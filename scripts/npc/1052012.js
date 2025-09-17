// Internet Cafe Access (Paid Entry)
var COST = 1000000;

npc.sendBackNext(
    "Hey there. Normally, only users logged in from an Internet Cafe can access the Premium Road.\r\n" +
    "If you're not at a cafe, I can still get you in... for a price.", false, true
);

if (npc.sendYesNo(
    "Want access to the Internet Cafe Premium Road right now for #b1,000,000 mesos#k?\r\n" +
    "Inside you'll find boosted EXP, drops, and mesos. Shall I send you?"
)) {
    if (plr.mesos() < COST) {
        npc.sendOk("You don't have enough mesos. Come back when you've got #b1,000,000#k.");
    } else {
        plr.giveMesos(-COST);
        plr.warp(193000000);
    }
} else {
    npc.sendBackNext("No problem. If you change your mind, I can grant access anytime for #b1,000,000 mesos#k.", false, true);
}
