// Regular Cab

var towns = [104000000, 103000000, 102000000, 101000000, 100000000]
var prices_num = [1000, 1000, 1000, 1000, 1000]

npc.sendNext("How's it going? I drive the Regular Cab. If you want to go from town to town safely and fast, then ride our cab. We'll gladly take you to your destination with an affordable price")

var text = "Choose your destination, for fees will change from place to place.\r\n"

var discountRate = (plr.job() == 0) ? 0.10 : 1.00

for (var i = 0; i  < towns.length; i++) {
    var cost = Math.floor(prices_num[i] * discountRate)
    text += "#L" + i + "##b#m" + towns[i] + "# (" + cost.toLocaleString() + " mesos)#l \r\n"
}

npc.sendSelection(text)

var sel = npc.selection()
var finalCost = Math.floor(prices_num[sel] * discountRate)

if (npc.sendYesNo("You don't have anything else to do here, huh? Do you really want to go to #b#m" + towns[sel] + "# #k? It'll cost you #b" + finalCost.toLocaleString() + " mesos")) {
    if (plr.mesos() < finalCost) {
        npc.sendOk("You don't have enough mesos! Come back when you do.")
    } else {
        plr.giveMesos(-1 * finalCost)
        plr.warp(towns[sel])
    }
} else {
    npc.sendNext("There's a lot to see in this town, too. Come back and find us when you need to go to a different town.")
}
