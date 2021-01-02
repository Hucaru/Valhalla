// Ellinia station boarding

function run(npc, player) {
    var inst = npc.getInstance(player)

    if ("canSellTickets" in inst.properties() && inst.properties()["canSellTickets"]) {
        npc.warpPlayer(player, 101000301)
    } else {
        npc.sendOK("Cannot board")
    }

    npc.terminate()
}