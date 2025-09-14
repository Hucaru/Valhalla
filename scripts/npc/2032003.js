npc.sendNext("How did you go through-such treacherous road to get here?? Incredible! #bThe Breath of Lava#k is here. Please give this to my brother. You'll finally be meeting up with the one you've been looking for, very soon.");

if (plr.itemCount(4031062) < 1) {
    npc.sendBackNext("Your etc, inventory seems to be full. Please make room in order to receive the item.", false, true);
} else {
    plr.giveEXP(15000);
    plr.giveItem(4031062, 1);
    plr.warp(211042300);
}