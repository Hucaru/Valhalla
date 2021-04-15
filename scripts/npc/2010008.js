// Lea guild npc

var state = 0

function run(npc, player) {
    if (npc.next()) {
        state++
    } else if (npc.back()) {
        state--
    }

    switch(state) {
    case 0:
        if (player.inGuild() && player.isGuildLeader()) {
            npc.sendYesNo("Would you like to update your guild emblem? This will cost 1,000,000 mesos.")
            state++
        } else {
            npc.sendOK("Please come back to me when you are a guild leader")
            npc.terminate()
        }
        break
    case 1:
        if (npc.yes()) {
            if (player.mesos() < 1e6) {
                npc.sendOK("You do not have enough mesos to change your emblem. Please come back when you have the required amount.")
            } else {
                npc.sendGuildEmblemEditor()
            }
        } else {
            npc.sendOK("Please come back when you wish to change your emblem.")
        }
        npc.terminate()
        break
    default:
        npc.terminate()
    }
}