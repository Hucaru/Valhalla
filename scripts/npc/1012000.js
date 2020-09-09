// Henesys Regular Cab

var towns = [104000000, 102000000, 101000000, 103000000]
var prices_text = ["800", "1,000", "1,000", "1,200"]
var prices_num = [800, 1000, 1000, 1200]

var state = 0
var selection = -1

function run(npc, player) {
    switch (state) {
    case 0:
        npc.sendBackNext("How's it going? I drive the Regular Cab. If you want to go from town to town safely and fast, then ride our cab. We'll gladly take you to your destination with an affordable price", false, true)
        state++ // state can only go forward from here, any other option terminates the npc chat
        break
    case 1:
        var text = "Choose your destination, for fees will change from place to place.\r\n"

        for (var i = 0; i  < towns.length; i++) {
            text += "#L" + i + "##b#m" + towns[i] + "# (" + prices_text[i] +" mesos)#l \r\n"
        }

        npc.sendSelection(text)
        state++ // state can only go forward from here, any other option terminates the npc chat
        break
    case 2:
        npc.sendYesNo("You don't have anything else to do here, huh? Do you really want to go to #b#m" + towns[npc.selection()] + "# #k? It'll cost you #b" + prices_text[npc.selection()] + " mesos")
        selection = npc.selection()
        state++
        break
    case 3:
        if (npc.yes()) {
            if (player.mesos() < prices_num[selection]) {
                npc.sendOK("You don't have enough mesos! Come back when you do.")
            } else if (npc.warpPlayer(player, towns[selection])) {
                player.giveMesos(-1 * prices_num[selection])
            } else {
                state = -1
            }
            npc.terminate()
        } else if (npc.no()) {
            npc.sendBackNext("There's a lot to see in this town, too. Come back and find us when you need to go to a different town.", false, true)
            npc.terminate()
        } else {
            state = -2
        }
        break
    default:
        npc.sendOK("Report this npc, it has entered unknown state: " + state + "selection:", selection)
        npc.terminate()
    }
    
}