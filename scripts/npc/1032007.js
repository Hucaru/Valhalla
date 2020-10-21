// Ellinia station ticket seller

function run(npc, player) {
    var inst = npc.getInstance(player)

    if ("canSellTickets" in inst.properties() && inst.properties()["canSellTickets"]) {
        npc.sendOK("Buy ticket")
        npc.terminate()
    } else {
        npc.sendOK("Cannot buy ticket")
        npc.terminate()
    }
}