// Orbis -> Ellinia station boarding

var props = plr.instanceProperties()

if ("canSellTickets" in props && props["canSellTickets"]) {
    plr.warp(200000112)
} else {
    npc.sendOk("Cannot board")
}