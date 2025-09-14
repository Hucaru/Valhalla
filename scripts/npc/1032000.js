// Regular Cab

var towns = [104000000, 102000000, 101000000, 103000000]
var prices_text = ["800", "1,000", "1,000", "1,200"]
var prices_num = [800, 1000, 1000, 1200]

npc.sendBackNext("How's it going? I drive the Regular Cab. If you want to go from town to town safely and fast, then ride our cab. We'll gladly take you to your destination with an affordable price", false, true)

var text = "Choose your destination, for fees will change from place to place.\r\n"

for (var i = 0; i  < towns.length; i++) {
    text += "#L" + i + "##b#m" + towns[i] + "# (" + prices_text[i] +" mesos)#l \r\n"
}

npc.sendSelection(text)

if (npc.sendYesNo("You don't have anything else to do here, huh? Do you really want to go to #b#m" + towns[npc.selection()] + "# #k? It'll cost you #b" + prices_text[npc.selection()] + " mesos")) {
    if (plr.mesos() < prices_num[npc.selection()]) {
        npc.sendOk("You don't have enough mesos! Come back when you do.")
    } else {
        plr.giveMesos(-1 * prices_num[npc.selection()])
        plr.warp(towns[npc.selection()])
    }
} else {
    npc.sendBackNext("There's a lot to see in this town, too. Come back and find us when you need to go to a different town.", false, true)
}