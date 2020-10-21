// Ellinia station boarding

function run(npc, player) {
    var inst = npc.getInstance(player)

    if ("canSellTickets" in inst.properties() && inst.properties()["canSellTickets"]) {
        npc.sendOK("Can board")
        npc.terminate()
    } else {
        npc.sendOK("Cannot board")
        npc.terminate()
    }
}