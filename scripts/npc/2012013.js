// Ticket Inspector at the pier
var hasWooden = plr.getSkillLevel(80001027) == 1
var hasRad = plr.getSkillLevel(80001028) == 1
var hasPlane = hasWooden || hasRad

if (!hasPlane) {
    // No airplane skill
    npc.sendSelection("Would you like to board the ship to Ludibrium? It take about 1 minute to arrive. \r\n#L0##bI'd like to board the ship.#l")
    var sel = npc.selection()
    if (sel == 0) {
        if (npc.sendYesNo("Would you like to board the ship to Ludibrium now?")) {
            plr.warp(200090100)
            plr.startMapTimeLimitTask(60, 220000110)
        } else {
            npc.sendBackNext("Do you have some business you need to take care of here?", true, true)
        }
    }
} else {
    // Has airplane skill
    var menu = "If you have an airplane, you can fly to stations all over the world. Would you rather take an airplane than wait for a ship? It'll cost you 5,000 mesos to use the station. \r\n\r\n"
    menu += "#b#L0#I'd like to use the plane. #r(5000 mesos)#l\r\n"
    menu += "#L1##bI'd like to board the ship.#l"
    npc.sendSelection(menu)
    var sel = npc.selection()

    if (sel == 0) {
        // Airplane choice
        var planeMenu = "Which airplane would you like to use? #b"
        if (hasWooden) planeMenu += "\r\n#L0#Wooden Airplane#l"
        if (hasRad) planeMenu += "\r\n#L1#Rad Airplane#l"
        npc.sendSelection(planeMenu)
        var planeSel = npc.selection()

        if (plr.mesos() >= 5000) {
            plr.takeMesos(5000)
            plr.giveBuff(planeSel == 0 ? 80001027 : 80001028, 1)
            plr.warp(200110021)
        } else {
            npc.sendOk("Please check and see if you have enough mesos to go.")
        }
    } else if (sel == 1) {
        // Ship choice
        if (npc.sendYesNo("Would you like to board the ship to Ludibrium now?")) {
            plr.warp(200090100)
            plr.startMapTimeLimitTask(60, 220000110)
        } else {
            npc.sendBackNext("Do you have some business you need to take care of here?", true, true)
        }
    }
}

// Generate by kimi-k2-instruct