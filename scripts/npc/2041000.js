```
if (plr.getSkillLevel(80001027) == 1 || plr.getSkillLevel(80001028) == 1) {
    var sel = npc.sendMenu(
        "If you have an airplane, you can fly to stations all over the world. Would you rather take an airplane than wait for a ship? It'll cost you 5,000 mesos to use the station.",
        "I'd like to use the plane. (5000 mesos)",
        "I'd like to board the ship."
    )

    if (sel == 0) {
        var planeIdx = 0
        if (plr.getSkillLevel(80001027) == 1 && plr.getSkillLevel(80001028) == 1) {
            planeIdx = npc.sendMenu(
                "Which airplane would you like to take?",
                "Wooden Airplane",
                "Rad Airplane"
            )
        }

        if (plr.mesos() > 5000) {
            plr.giveMesos(-5000)
            plr.giveBuff(planeIdx == 0 ? 80001027 : 80001028, 1)
            plr.warp(200110020)
        } else {
            npc.sendOk("You don't have enough money for the Station fee.")
        }
    } else if (sel == 1) {
        if (npc.sendYesNo("Would you like to board the ship to Orbis now? It takes about a minute to get there.")) {
            plr.warp(200090110)
        }
    }
} else {
    if (npc.sendYesNo("Would you like to board the ship to Orbis now? It takes about a minute to get there.")) {
        plr.warp(200090110)
    }
}
```