// Orbis station ticket seller

var props = plr.instanceProperties()

if ("canSellTickets" in props && props["canSellTickets"]) {
    npc.sendOk("Buy ticket")
} else {
    npc.sendOk("Cannot buy ticket")
}