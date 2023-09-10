// Ellinia station boarding

var props = plr.instanceProperties()

if ("canSellTickets" in props && props["canSellTickets"]) {
    plr.warp(101000301)
} else {
    npc.sendOk("Cannot board")
}