npc.sendBackNext("Oceans are all connected to each other. Places you can't reach by foot can be easily reached oversea. How about taking #bDolphin Taxi#k with us today?", false, true)

var menu = ""
if (plr.mapId() == 230000000) {
    menu += "#L0#Go to the Sharp Unknown. (Towards Ludibrium/Korean Folk Town)#l\r\n"
    menu += "#L1#Go to Herb Town.#l"
} else {
    menu += "#L1#Go to Acuariurm.#l"
}
npc.sendSelection(menu)
var sel = npc.selection()

var fare = (sel < 1 ? 1000 : 10000)
if (npc.sendYesNo("The fare is " + fare + " mesos. Shall we go?")) {
    if (plr.mesos() < fare) {
        npc.sendOk("You don't have enough mesos.")
    } else {
        plr.takeMesos(fare)
        var destMap = (sel < 1 ? 230030200 : (plr.mapId() == 251000100 ? 230000000 : 251000100))
        plr.warp(destMap)
    }
} else {
    npc.sendOk("Okay, next time.")
}

// Generate by kimi-k2-instruct