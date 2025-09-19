// Fairy Master Alchemist

npc.sendBackNext(
    "Yeah... I am the master alchemist of the fairies. But the fairies are not supposed to be in contact with a human being for a long period of time... a strong person like you will be fine, though. If you get me the materials, I'll make you a special item.",
    false,
    true
);

var sel = npc.sendMenu(
    "What do you want to make?",
    "Moon Rock",
    "Star Rock",
    "Black Feather"
);

if (sel === 0) {
    // Moon Rock: requires 1x each 4011000..4011006 and 10,000 mesos -> 4011007
    if (npc.sendYesNo(
        "So you want to make #b#t4011007##k?\r\n" +
        "To do that you need #b1 of each#k refined plate:\r\n" +
        "#t4011000#, #t4011001#, #t4011002#, #t4011003#, #t4011004#, #t4011005#, #t4011006#.\r\n" +
        "Throw in #b10,000 mesos#k and I'll make it for you."
    )) {
        var ok =
            plr.itemCount(4011000) && plr.itemCount(4011001) && plr.itemCount(4011002) &&
            plr.itemCount(4011003) && plr.itemCount(4011004) && plr.itemCount(4011005) &&
            plr.itemCount(4011006) && plr.mesos() >= 10000;

        if (ok) {
            // Remove materials
            plr.removeItemsByID(4011000, 1); plr.removeItemsByID(4011001, 1); plr.removeItemsByID(4011002, 1);
            plr.removeItemsByID(4011003, 1); plr.removeItemsByID(4011004, 1); plr.removeItemsByID(4011005, 1);
            plr.removeItemsByID(4011006, 1);
            plr.takeMesos(10000);

            // Give result
            if (plr.giveItem(4011007, 1)) {
                npc.sendOk("Okay, here—#b#t4011007##k. It's well-made, probably because I'm using good materials. If you need my help down the road, feel free to come back.");
            } else {
                npc.sendOk("Please make room in your Etc/Use inventory before crafting.");
            }
        } else {
            npc.sendOk("Are you sure you have enough mesos and the refined plates?\r\nYou need #b1 of each#k: #t4011000#, #t4011001#, #t4011002#, #t4011003#, #t4011004#, #t4011005#, #t4011006# and #b10,000 mesos#k.");
        }
    }
} else if (sel === 1) {
    // Star Rock: requires 1x each 4021000..4021008 and 15,000 mesos -> 4021009
    if (npc.sendYesNo(
        "So you want to make #b#t4021009##k?\r\n" +
        "To do that you need #b1 of each#k refined gemstone:\r\n" +
        "#t4021000#, #t4021001#, #t4021002#, #t4021003#, #t4021004#, #t4021005#, #t4021006#, #t4021007#, #t4021008#.\r\n" +
        "Throw in #b15,000 mesos#k and I'll make it for you."
    )) {
        var haveAll = true;
        for (var g = 4021000; g <= 4021008; g++) {
            if (!plr.itemCount(g)) { haveAll = false; break; }
        }
        var ok2 = haveAll && plr.mesos() >= 15000;

        if (ok2) {
            for (var r = 4021000; r <= 4021008; r++) plr.removeItemsByID(r, 1);
            plr.takeMesos(15000);

            if (plr.giveItem(4021009, 1)) {
                npc.sendOk("Okay, here—#b#t4021009##k. It's well-made, probably because I'm using good materials. If you need my help down the road, feel free to come back.");
            } else {
                npc.sendOk("Please make room in your Etc/Use inventory before crafting.");
            }
        } else {
            npc.sendOk("Are you sure you have enough mesos and the gemstones?\r\nYou need #b1 of each#k: #t4021000# ~ #t4021008# and #b15,000 mesos#k.");
        }
    }
} else if (sel === 2) {
    // Black Feather: requires 4001006, 4011007, 4021008 and 30,000 mesos -> 4031042
    if (npc.sendYesNo(
        "So you want to make #b#t4031042##k?\r\n" +
        "To do that you need #b1 #t4001006#, 1 #t4011007#, and 1 #t4021008##k.\r\n" +
        "Throw in #b30,000 mesos#k and I'll make it for you.\r\n\r\n" +
        "Note: This feather is special; dropping it may make it disappear, and it cannot be traded."
    )) {
        var ok3 =
            plr.itemCount(4001006) &&
            plr.itemCount(4011007) &&
            plr.itemCount(4021008) &&
            plr.mesos() >= 30000;

        if (ok3) {
            plr.removeItemsByID(4001006, 1);
            plr.removeItemsByID(4011007, 1);
            plr.removeItemsByID(4021008, 1);
            plr.takeMesos(30000);

            if (plr.giveItem(4031042, 1)) {
                npc.sendOk("Okay, here—#b#t4031042##k. It's well-made, probably because I'm using good materials. If you need my help down the road, feel free to come back.");
            } else {
                npc.sendOk("Please make room in your Etc/Use inventory before crafting.");
            }
        } else {
            npc.sendOk("Are you sure you have enough mesos and the materials?\r\nYou need #b1 #t4001006#, 1 #t4011007#, 1 #t4021008##k and #b30,000 mesos#k.");
        }
    }
}