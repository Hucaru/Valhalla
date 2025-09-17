if (plr.itemCount(4031013) >= 30) {
    npc.sendNext("Ohhhhh.. you collected all 30 Dark Marbles!! It should have been difficult.. just incredible! Alright. You've passed the test and for that, I'll reward you #bThe Proof of a Hero#k. Take that and go back to Ellinia.");
    plr.warp(101020000);
    plr.removeItemsByID(4031013, 9999);
    plr.giveItem(4031009, -1);
    plr.giveItem(4031012, 1);
} else {
    npc.sendOk("You will have to collect me #b30 #t4031013##k. Good luck.");
}