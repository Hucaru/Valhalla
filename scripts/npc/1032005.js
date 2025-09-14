npc.sendNext("Hi there! This cab is for VIP customers only. Instead of just taking you to different towns like the regular cabs, we offer a much better service worthy of VIP class. It's a bit pricey, but... for only 10,000 mesos, we'll take you safely to #bAnt Tunnel#k.")

let fee = 10000;
if (plr.job() === 0) {
    // Beginner discount
    fee = 1000;
    if (npc.sendYesNo("There's a special 90% discount for all beginners. Ant Tunnel is located deep inside in the dungeon that's at the center of the Victoria Island, where the 24 Hr Mobile Store is. Would you like to go there for #b1,000 mesos#k?")) {
        if (plr.mesos() < 1000) {
            npc.sendOk("It looks like you don't have enough mesos. Sorry but you won't be able to use this without it.")
        } else {
            plr.giveMesos(-1000)
            plr.warp(105070001)
        }
    } else {
        npc.sendOk("This town also has a lot to offer. Find us if and when you feel the need to go to the Ant Tunnel Park.")
    }
} else {
    if (npc.sendYesNo("Ant Tunnel is located deep inside in the dungeon that's at the center of the Victoria Island, where the 24 Hr Mobile Store is. Would you like to go there for #b10,000 mesos#k?")) {
        if (plr.mesos() < 10000) {
            npc.sendOk("It looks like you don't have enough mesos. Sorry but you won't be able to use this without it.")
        } else {
            plr.giveMesos(-10000)
            plr.warp(105070001)
        }
    } else {
        npc.sendOk("This town also has a lot to offer. Find us if and when you feel the need to go to the Ant Tunnel Park.")
    }
}