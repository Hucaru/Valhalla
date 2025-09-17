// Internet-Cafe Premium Road warper

var maps = [190000000, 191000000, 192000000, 195000000];

npc.sendBackNext(
    "Bzzzt~Beep~Boop!! Welcome.. to the Internet Cafe!!.. I.. can warp.. you.. to different.. training areas.. within.. the special.. Internet Cafe.. Premium Road.. areas.... Experience points.. from.. monsters.. are doubled.. as well as.. drops.. and.. mesos!!..",
    false, true
);

var text = "Please.. choose.. a.. destination....\r\n";
for (var i = 0; i < maps.length; i++) {
    text += "#L" + i + "##m" + maps[i] + "##l\r\n";
}

npc.sendSelection(text);

var sel = npc.selection();
if (sel < 0 || sel >= maps.length) {
    npc.sendOk("Bzzzt.. invalid.. selection..");
} else {
    plr.warp(maps[sel]);
}