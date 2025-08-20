npc.sendSelection("It's a magic stone for Eos Tower tourists. It will take you to your desired location for a small fee. \r\n(You can use a #bEos Rock Scroll#k in lieu of mesos.)\r\n#L0##b#m221020000# (15000 Mesos)#l\r\n#L1##b#m221021200# (15000 Mesos)#l\r\n#L2##b#m221023200# (15000 Mesos)#l")
var select = npc.selection()

if (npc.sendYesNo("Would you like to move to #b#m" + [221020000, 221021200, 221023200][select] + "##k? The price is #b15000 mesos#k.")) {
    if (plr.mesos() < 15000) {
        npc.sendOk("You don't have enough mesos. Sorry, but you can't use this service if you can't pay the fee.")
    } else {
        plr.takeMesos(15000)
        plr.warp([221020000, 221021200, 221023200][select])
    }
} else {
    npc.sendOk("Please try again later.")
}

// Generate by kimi-k2-instruct