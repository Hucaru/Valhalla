// Orbis -> Ludi station boarding

var props = plr.instanceProperties()

if ("canSellTickets" in props && props["canSellTickets"]) {
    plr.warp(200000122)
} else {
    npc.sendOk("Cannot board")
}