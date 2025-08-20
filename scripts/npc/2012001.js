// Check if player has airplane skill
var hasAirplane = (plr.getSkillLevel(80001027) == 1 || plr.getSkillLevel(80001028) == 1)

if (!hasAirplane) {
    // No airplane – only ship option
    if (npc.sendYesNo("Would you like to board the ship for #bVictoria Island#k now? It'll take about 30 seconds to get there.")) {
        plr.warp(200090000)
    } else {
        npc.sendOk("Do you have some business you need to take care of here?")
    }
} else {
    // Has airplane – show airplane vs ship choice
    var menu = "If you have an airplane, you can fly to stations all over the world. Would you rather take an airplane than wait for a ship? It'll cost you 5000 mesos. \r\n\r\n"
    menu += "#L0#Use the airplane. #r(5000 mesos)#l\r\n"
    menu += "#L1#Board a ship.#l"
    npc.sendSelection(menu)
    var choice = npc.selection()

    if (choice == 0) {
        // Airplane selection
        var planeMenu = "Which airplane would you like to use? #b"
        if (plr.getSkillLevel(80001027) == 1) {
            planeMenu += "\r\n#L0#Wooden Airplane#l"
        }
        if (plr.getSkillLevel(80001028) == 1) {
            planeMenu += "\r\n#L1#Rad Airplane#l"
        }
        npc.sendSelection(planeMenu)
        var planeChoice = npc.selection()

        if (plr.mesos() >= 5000) {
            plr.takeMesos(5000)
            plr.giveBuff(planeChoice == 0 ? 80001027 : 80001028, 1)
            plr.warp(200110001)
        } else {
            npc.sendOk("You don't have enough money for the Station fee.")
        }
    } else if (choice == 1) {
        // Ship option
        if (npc.sendYesNo("Would you like to board the ship for #bVictoria Island#k now? It'll take about 30 seconds to get there.")) {
            plr.warp(200090000)
        } else {
            npc.sendOk("Do you have some business you need to take care of here?")
        }
    }
}

// Generate by kimi-k2-instruct