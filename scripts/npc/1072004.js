npc.sendBackNext("You will have to collect me #b30 #t4031013##k. Good luck.", false, true);

if (plr.itemCount(4031013) >= 30) {
    npc.sendBackNext("Ohhhhh.. you collected all 30 Dark Marbles!! It should have been difficult.. just incredible! Alright. You've passed the test and for that, I'll reward you #bThe Proof of a Hero#k. Take that and go back to Perion.", true, true);
    plr.warp(102020300, 0);
    plr.removeItemsByID(4031013, 30);
    plr.takeItem(4031008, 0, 1, 1);   // ETC inv
    plr.giveItem(4031012, 1);
}