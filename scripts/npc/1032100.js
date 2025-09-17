npc.sendBackNext("Yeah... I am the master alchemist of the fairies. But the fairies are not supposed to be in contact with a human being for a long period of time... a strong person like you will be fine, though. If you get me the materials, I'll make you a special item.", false, true)

var sel = npc.sendMenu("What do you want to make?", "Moon Rock", "Star Rock", "Black Feather");

var item, neededMeso, materials;

if (sel === 0) {
    item = "Moon Rock";
    neededMeso = 10000;
    if (npc.sendYesNo("So you want to make Moon Rock To do that you need refined one of each of these: #bBronze Plate, Steel Plate, Mithril Plate, Adamantium Plate, Silver, Orihalcon Plate and Gold#k. Throw in 10,000 mesos and I'll make it for you.")) {
        if (plr.itemCount(4011000) && plr.itemCount(4011001) && plr.itemCount(4011002) && plr.itemCount(4011003) && plr.itemCount(4011004) && plr.itemCount(4011005) && plr.itemCount(4011006) && plr.mesos() >= 10000) {
            plr.removeItemsByID(4011000, 1); plr.removeItemsByID(4011001, 1); plr.removeItemsByID(4011002, 1);
            plr.removeItemsByID(4011003, 1); plr.removeItemsByID(4011004, 1); plr.removeItemsByID(4011005, 1);
            plr.removeItemsByID(4011006, 1);
            plr.takeMesos(10000);
            plr.giveItem(4011007, 1);
            npc.sendOk("Ok here, take Moon Rock. It's well-made, probably because I'm using good materials. If you need my help down the road, feel free to come back.");
        } else {
            npc.sendOk("Are you sure you have enough mesos? Please check and see if you have the refined #bBronze Plate, Steel Plate, Mithril Plate, Adamantium Plate, Silver, Orihalcon Plate and Gold#k, one of each.");
        }
    }
} else if (sel === 1) {
    item = "Star Rock";
    neededMeso = 15000;
    if (npc.sendYesNo("So you want to make the Star Rock? To do that you need refined one of each of these: #bGarnet, Amethyst, AquaMarine, Emerald, Opal, Sapphire, Topaz, Diamond and Black Crystal#k. Throw in 15,000 mesos and I'll make it for you.")) {
        if (plr.itemCount(4021000) && plr.itemCount(4021001) && plr.itemCount(4021002) && plr.itemCount(4021003) &&
            plr.itemCount(4021004) && plr.itemCount(4021005) && plr.itemCount(4021006) && plr.itemCount(4021007) && plr.itemCount(4021008) && plr.mesos() >= 15000) {
            for (var i = 4021000; i <= 4021008; i++) plr.removeItemsByID(i, 1);
            plr.takeMesos(15000);
            plr.giveItem(4021009, 1);
            npc.sendOk("Ok here, take Star Rock. It's well-made, probably because I'm using good materials. If you need my help down the road, feel free to come back.");
        } else {
            npc.sendOk("Are you sure you have enough mesos? Please check and see if you have the refined #bGarnet, Amethyst, AquaMarine, Emerald, Opal, Sapphire, Topaz, Diamond, Black Crystal#k, one of each.");
        }
    }
} else if (sel === 2) {
    item = "Black Feather";
    neededMeso = 30000;
    if (npc.sendYesNo("So you want to make Black Feather To do that you need #b1 Flaming Feather, 1 Moon Rock and 1 Black Crystal#k. Throw in 30,000 mesos and I'll make it for you. Oh yeah, this piece of feather is a very special item, so if you drop it by any chance, it'll disappear, as well as you won't be able to give it away to someone else.")) {
        if (plr.itemCount(4001006) && plr.itemCount(4011007) && plr.itemCount(4021008) && plr.mesos() >= 30000) {
            plr.removeItemsByID(4001006, 1); plr.removeItemsByID(4011007, 1); plr.removeItemsByID(4021008, 1);
            plr.takeMesos(30000);
            plr.giveItem(4031042, 1);
            npc.sendOk("Ok here, take Black Feather. It's well-made, probably because I'm using good materials. If you need my help down the road, feel free to come back.");
        } else {
            npc.sendOk("Are you sure you have enough mesos? Check and see if you have #b1 Flaming Feather, #b1 Moon Rock and 1 Black Crystal#k ready for me.");
        }
    }
}