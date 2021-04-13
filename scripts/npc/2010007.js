// Heracle guild npc
var state = 0

function run(npc, player) {
    if (npc.next()) {
        state++
    } else if (npc.back()) {
        state--
    }

    switch(state) {
        case 0:
            npc.sendBackNext("Hey...are you interested in GUILDS by any chance?", false, true)
            break
        case 1:
            npc.sendSelection("#b#L0#What's a Guild#l\r\n#L1#What do I do to form a guild?#l\r\n#L2#I want to start a guild#l")
            state++
            break
        case 2:
            var selection = npc.selection()

            if (selection == 0) {
                npc.sendBackNext("A guild is....", true, false)
            } else if (selection == 1) {
                npc.sendBackNext("In order to form a guild you need to be in a party of 5 people and the party leader and have at least 5,000,000 mesos", true, false)
            } else if (selection == 2) {
                npc.sendYesNo("Are you sure wish to form a guild?")
                state++
            } else {
                npc.sendOK("Unknown selection: " + npc.selection())
                npc.terminate()
            }

            break
        case 3:
            if (npc.yes()) {
                npc.startGuildCreation(player)
            } else {
                npc.sendOK("Come back when you have decided to form a guild")
            }

            npc.terminate()
            break
        default:
            npc.sendOK("state " + state)
            npc.terminate()
    }
}