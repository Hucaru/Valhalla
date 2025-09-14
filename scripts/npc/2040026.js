var maps = [221020000, 221022100, 221023200];   // destinations

npc.sendNext("It's a magic stone for Eos Tower tourists. It will take you to your desired location for a small fee. \r\n(You can use a #bEos Rock Scroll#k in lieu of mesos.)");

var choices = "";
for (var i = 0; i < maps.length; i++) {
    choices += "\r\n#L" + i + "##b#m" + maps[i] + "# (15000 Mesos)#l";
}

// Ask which map to choose
var sel = npc.askMenu("Select your destination:", choices);

if (npc.sendYesNo("Would you like to move to #b#m" + maps[sel] + "##k? The price is #b15000 mesos#k.")) {
    if (plr.mesos() < 15000) {
        npc.sendOk("You don't have enough mesos. Sorry, but you can't use this service if you can't pay the fee.");
    } else {
        plr.giveMesos(-15000);
        plr.warp(maps[sel]);
    }
}