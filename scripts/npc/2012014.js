var mapArray = [200080600, 200081400, 200082100];

npc.sendSelection(
    "This is a magic stone for Orbis Tower travelers. This will take you to your desired floor if you pay a fee. \r\n" +
    "(You can use the #bOrbis Rock Scroll#k in place of Mesos.)\r\n" +
    "#L0##b#m200080600 # (5000 Mesos)#l\r\n" +
    "#L1##b#m200081400 # (5000 Mesos)#l\r\n" +
    "#L2##b#m200082100 # (5000 Mesos)#l"
);

var select = npc.selection();

if (npc.sendYesNo("Would you like to move to #b#m" + mapArray[select] + "##k? The price is #b5000 mesos#k.")) {
    if (plr.mesos() < 5000) {
        npc.sendOk("You don't have enough mesos. Sorry, but you can't use this service if you can't pay the fee.");
    } else {
        plr.giveMesos(-5000);
        plr.warp(mapArray[select]);
    }
}