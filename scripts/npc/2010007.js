// Heracle guild npc

if (plr.inGuild() && plr.guildRank() == 1) {
    npc.sendSelection("What would you like to do:\r\n#b#L0#Increase guild capacity#l\r\n#L1#Disband guild#l")

    switch(npc.selection()) {
        case 0:
            // increase capacity
            break
        case 1:
            if (npc.sendYesNo("Are you sure that you wish to disband the guild?")) {
                npc.sendOk("The guild has been disbanded") // I would have though the client would have this pre-baked like other guild management messages
                plr.disbandGuild()
            } else {
                npc.sendOk("Please come back when you wish to disband the guild.")
            }
            break
    }
} else {
    npc.sendBackNext("Hey...are you interested in GUILDS by any chance?", false, true)
    npc.sendSelection("#b#L0#What's a Guild#l\r\n#L1#What do I do to form a guild?#l\r\n#L2#I want to start a guild#l")
    
    switch(npc.selection()) {
        case 0:
            npc.sendBackNext("A guild is....", true, false)
            break;
        case 1:
            npc.sendBackNext("In order to form a guild you need to be in a party of 5 people and the party leader and have at least 5,000,000 mesos.", true, false)
            break;
        case 2:
            if (npc.sendYesNo("Are you sure wish to form a guild?")) {
                if (plr.inGuild()) {
                    npc.sendOk("You cannot create a guild whilst still being in one.")
                } else if (!plr.inParty()) {
                    npc.sendOk("In order to form a guild you need to be in a party with at least 5 people and all be in this room.")
                } else if (!plr.isPartyLeader()) {
                    npc.sendOk("You must be the party leader in order to form a guild.")
                } else if (plr.partyMembersOnMapCount() < 2) {
                    npc.sendOk("You must have 4 other people in this room in order to form a guild.")
                } else if (plr.mesos() < 5e6) {
                    npc.sendOk("You need at least 5,000,000 mesos to form a guild. Please come back once you have the required amount.")
                } else {
                    npc.sendGuildCreation()
                }
            } else {
                npc.sendOk("Come back when you have decided to form a guild.")
            }
            break
        default:
            npc.sendOk("Error in selection")
    }
}