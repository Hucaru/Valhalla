// It's a magic stone for Eos Tower tourists
var map = [221020000, 221021200, 221022100]
var str = "It's a magic stone for Eos Tower tourists. It will take you to your desired location for a small fee. \r\n(You can use a #bEos Rock Scroll#k in lieu of mesos.)"
for (var i = 0; i < map.length; i++) {
    str += "\r\n#L" + i + "##b#m" + map[i] + "#(15000 Mesos)#l"
}
npc.sendSelection(str)
var select = npc.selection()

if (npc.sendYesNo("Would you like to move to #b#m" + map[select] + "##k? The price is #b15000 mesos#k.")) {
    if (plr.mesos() < 15000) {
        npc.sendOk("You don't have enough mesos. Sorry, but you can't use this service if you can't pay the fee.")
    } else {
        plr.takeMesos(15000)
        plr.warp(map[select])
    }
} else {
    npc.sendOk("Please try again later.")
}
// Generate by kimi-k2-instruct