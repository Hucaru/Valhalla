// Ticket Inspector at the pier
var hasWooden = (plr.getSkillLevel(80001027) == 1)
var hasRad = (plr.getSkillLevel(80001028) == 1)

if (!hasWooden && !hasRad) {
    // No airplane skills – only ship option
    if (npc.sendYesNo("Would you like to board the ship to Orbis now? It takes about a minute to get there.")) {
        plr.warp(200090110)
    } else {
        npc.sendOk("Do you have some business you need to take care of here?")
    }
} else {
    // Has airplane skills – show menu
    var menu = "If you have an airplane, you can fly to stations all over the world. Would you rather take an airplane than wait for a ship? It'll cost you 5,000 mesos to use the station.\r\n\r\n"
    menu += "#L0#I'd like to use the plane. #r(5000 mesos)#l\r\n"
    menu += "#L1#I'd like to board the ship.#l"
    npc.sendSelection(menu)
    var choice = npc.selection()

    if (choice == 0) {
        // Airplane selection
        var planeMenu = "Which airplane would you like to take?#b"
        if (hasWooden) planeMenu += "\r\n#L0#Wooden Airplane#l"
        if (hasRad) planeMenu += "\r\n#L1#Rad Airplane#l"
        npc.sendSelection(planeMenu)
        var planeChoice = npc.selection()

        if (plr.mesos() >= 5000) {
            plr.takeMesos(5000)
            plr.giveBuff(planeChoice == 0 ? 80001027 : 80001028, 1)
            plr.warp(200110020)
        } else {
            npc.sendOk("You don't have enough money for the Station fee.")
        }
    } else if (choice == 1) {
        // Ship option
        if (npc.sendYesNo("Would you like to board the ship to Orbis now? It takes about a minute to get there.")) {
            plr.warp(200090110)
        } else {
            npc.sendOk("Do you have some business you need to take care of here?")
        }
    }
}

// Generate by kimi-k2-instruct