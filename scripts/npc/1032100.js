npc.sendBackNext("Yeah... I am the master alchemist of the fairies. But the fairies are not supposed to be in contact with a human being for a long period of time... a strong person like you will be fine, though. If you get me the materials, I'll make you a special item.", false, true)

var menu = "What do you want to make? #b\r\n#L0#Moon Rock#l\r\n#L1#Star Rock#l\r\n#L2#Black Feather#l"
npc.sendSelection(menu)
var select = npc.selection()

var item, cost, needed, ok = false

if (select == 0) {
    item = "Moon Rock"
    cost = 10000
    needed = [4011000, 4011001, 4011002, 4011003, 4011004, 4011005, 4011006]
    if (npc.sendYesNo("So you want to make Moon Rock To do that you need refined one of each of these: #bBronze Plate, Steel Plate, Mithril Plate, Adamantium Plate, Silver, Orihalcon Plate and Gold#k. Throw in 10,000 mesos and I'll make it for you.")) {
        ok = true
    }
} else if (select == 1) {
    item = "Star Rock"
    cost = 15000
    needed = [4021000, 4021001, 4021002, 4021003, 4021004, 4021005, 4021006, 4021007, 4021008]
    if (npc.sendYesNo("So you want to make the Star Rock? To do that you need refined one of each of these: #bGarnet, Amethyst, AquaMarine, Emerald, Opal, Sapphire, Topaz, Diamond and Black Crystal#k. Throw in 15,000 mesos and I'll make it for you.")) {
        ok = true
    }
} else if (select == 2) {
    item = "Black Feather"
    cost = 30000
    needed = [4001006, 4011007, 4021008]
    if (npc.sendYesNo("So you want to make Black Feather To do that you need #b1 Flaming Feather, 1 Moon Rock and 1 Black Crystal#k. Throw in 30,000 mesos and I'll make it for you. Oh yeah, this piece of feather is a very special item, so if you drop it by any chance, it'll disappear, as well as you won't be able to give it away to someone else.")) {
        ok = true
    }
}

if (ok) {
    var hasAll = true
    for (var i = 0; i < needed.length; i++) {
        if (plr.itemQuantity(needed[i]) < 1) {
            hasAll = false
            break
        }
    }
    if (hasAll && plr.mesos() >= cost) {
        for (var i = 0; i < needed.length; i++) {
            plr.takeItem(needed[i], 1)
        }
        plr.takeMesos(cost)
        if (select == 0) {
            plr.giveItem(4011007, 1)
        } else if (select == 1) {
            plr.giveItem(4021009, 1)
        } else if (select == 2) {
            plr.giveItem(4031042, 1)
        }
        npc.sendBackNext("Ok here, take " + item + ". It's well-made, probably because I'm using good materials. If you need my help down the road, feel free to come back.", true, true)
    } else {
        if (select == 0) {
            npc.sendBackNext("Are you sure you have enough mesos? Please check and see if you have the refined #bBronze Plate, Steel Plate, Mithril Plate, Adamantium Plate, Silver, Orihalcon Plate and Gold#k, one of each.", true, true)
        } else if (select == 1) {
            npc.sendBackNext("Are you sure you have enough mesos? Please check and see if you have the refined #bGarnet, Amethyst, AquaMarine, Emerald, Opal, Sapphire, Topaz, Diamond, Black Crystal#k, one of each.", true, true)
        } else if (select == 2) {
            npc.sendBackNext("Are you sure you have enough mesos? Check and see if you have #b1 Flaming Feather, #b1 Moon Rock