// Heracle guild npc
var state = 0

var manageSelection = -1

function run(npc, player) {
    if (npc.next()) {
        state++
    } else if (npc.back()) {
        state--
    }

    switch(state) {
        case 0:
            if (player.inGuild() && player.isGuildLeader()) {
                npc.sendSelection("What would you like to do:\r\n#b#L0#Increase guild capacity#l\r\n#L1#Disband guild#l")
                state = 4
            } else {
                npc.sendBackNext("Hey...are you interested in GUILDS by any chance?", false, true)
            }
            
            break
        /*
         * Guild Creation
         */
        case 1:
            npc.sendSelection("#b#L0#What's a Guild#l\r\n#L1#What do I do to form a guild?#l\r\n#L2#I want to start a guild#l")
            state++
            break
        case 2:
            var selection = npc.selection()

            if (selection == 0) {
                npc.sendBackNext("A guild is....", true, false)
            } else if (selection == 1) {
                npc.sendBackNext("In order to form a guild you need to be in a party of 5 people and the party leader and have at least 5,000,000 mesos.", true, false)
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
                if (player.inGuild()) {
                    npc.sendOK("You cannot create a guild whilst still being in one.")
                } else if (!player.inParty()) {
                    npc.sendOK("In order to form a guild you need to be in a party with at least 5 people and all be in this room.")
                } else if (!player.isPartyLeader()) {
                    npc.sendOK("You must be the party leader in order to form a guild.")
                } else if (player.partyMembersOnMapCount() < 5) {
                    npc.sendOK("You must have 4 other people in this room in order to form a guild.")
                } else if (player.mesos() < 5e6) {
                    npc.sendOK("You need at least 5,000,000 mesos to form a guild. Please come back once you have the required amount.")
                } else {
                    npc.sendGuildCreation()
                }
            } else {
                npc.sendOK("Come back when you have decided to form a guild.")
            }

            npc.terminate()
            break
        /*
         * Manage Guild
         */
        case 4:
            manageSelection = npc.selection()

            if (manageSelection == 0) {
                // increase capacity
            } else if (manageSelection == 1) {
                npc.sendYesNo("Are you sure that you wish to disband the guild?")
            } else {
                npc.sendOK("Bug in state 4")
                npc.terminate()
            }

            state++
            break

        case 5:
            if (npc.yes()) {
                if (manageSelection == 0) {
                    // increase capacity
                } else if (manageSelection == 1) {
                    npc.sendOK("The guild has been disbanded") // I would have though the client would have this pre-baked like other guild management messages
                    player.disbandGuild()
                }
            } else {
                if (manageSelection == 0) {
                    npc.sendOK("Please come back when you have decided to increase your guild capacity.")
                } else if (manageSelection == 1) {
                    npc.sendOK("Please come back when you wish to disband the guild.")
                }
            }
            
            npc.terminate()
            break
        default:
            npc.sendOK("state " + state)
            npc.terminate()
    }
}