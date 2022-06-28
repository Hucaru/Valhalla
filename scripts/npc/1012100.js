// Athena Pierce
var state = 0
var selection = -1

function run(npc, player) {
    switch (state) {
    case 0:
        if ( player.job() == 0 ) {
            if ( player.level() >= 10 && player.job() == 0) {
                npc.sendBackNext("So you decided to become a #rBowman#k?");
                } else {
                npc.sendOK("Train a bit more and I can show you the way of the #rBowman#k.")
                npc.terminate();
                }
        } else {
            npc.sendOK("The progress you have made is astonishing.")
            npc.terminate(); 
        }
        state++ // state can only go forward from here, any other option terminates the npc chat
        break
    case 1:
        npc.sendBackNext("It is an important and final choice. You will not be able to turn back.", false, true)
        state++ // state can only go forward from here, any other option terminates the npc chat
        break
    case 2:
        npc.sendYesNo("Do you want to become a #rBowman#k?")
        selection = npc.selection()
        state++
        break
    case 3:
        if (npc.yes()) {
            player.giveJob(300)
            player.gainItem(1452002, 1)
            npc.sendOK("So be it! Now go, and go with pride.")
            npc.terminate()
        } else if (npc.no()) {
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