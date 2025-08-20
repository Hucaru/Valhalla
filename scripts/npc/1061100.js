npc.sendBackNext("Welcome. We're the #m105000000# Hotel. Our hotel works hard to serve you the best at all times. If you are tired and worn out from hunting, how about a relaxing stay at our hotel?", false, true)

var map = [105000011, 105000012]
var cost = [499, 999]

var chat = "We offer two kinds of rooms for service. Please choose the one of your liking."
for (var i = 0; i < map.length; i++) {
    chat += "\r\n#L" + i + "##b#m" + map[i] + "#(" + cost[i] + " Mesos per use)#l"
}
npc.sendSelection(chat)
var select = npc.selection()

var confirmMsg = select < 1
    ? "You've chosen the regular sauna. Your HP and MP will recover fast and you can even purchase some items there. Are you sure you want to go in?"
    : "You've chosen the VIP sauna. your HP and MP will recover even faster than that of the regular sauna and you can even find a special item in there. Are you sure you want to go in?"

if (npc.sendYesNo(confirmMsg)) {
    if (plr.mesos() < cost[select]) {
        npc.sendBackNext("I'm sorry. It looks like you don't have mesos. It will cost you " + (select < 1 ? "at least " : "") + "" + cost[select] + " Mesos to stay at our hotel.", true, true)
    } else {
        plr.takeMesos(cost[select])
        plr.warp(map[select])
    }
} else {
    npc.sendBackNext("We offer other kinds of services, too, so please think carefully and then make your decision.", true, true)
}

// Generate by kimi-k2-instruct