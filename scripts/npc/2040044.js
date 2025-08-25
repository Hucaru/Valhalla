var eim = plr.instanceProperties();
if (eim["stage9"] == null) {
    npc.sendOk("Now that you've come this far, it's time to defeat the one responsible for this mess, #b#o9300012##k. I suggest you be careful, though, as he is not in a very good mood. If you and your party members defeat him, the Dimensional Schism will close forever. I'm counting on you!");
} else {
    npc.sendBackNext("You've defeated #b#o9300012##k! Magnificent! Thanks to you, the Dimensional Schism has been safely closed. I will now help you leave this place.", false, true);
    plr.warp(922011100);
}

// Generate by kimi-k2-instruct