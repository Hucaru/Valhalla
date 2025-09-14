// 紫氣球 – Warp only after stage9 is set
var inst = instanceProperties();

if (inst.stage9 === null) {
    npc.sendNext("Now that you've come this far, it's time to defeat the one responsible for this mess, #b#o9300012##k. I suggest you be careful, though, as he is not in a very good mood. If you and your party members defeat him, the Dimensional Schism will close forever. I'm counting on you!");
} else {
    npc.sendNext("You've defeated #b#o9300012##k! Magnificent! Thanks to you, the Dimensional Schism has been safely closed. I will now help you leave this place.");
    plr.warp(922011100);
}