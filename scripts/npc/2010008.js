// Lea guild npc

if (plr.inGuild() && plr.guildRank() == 1) {
    if (npc.sendYesNo("Would you like to update your guild emblem? This will cost 1,000,000 mesos.")) {
        if (plr.mesos() < 1e6) {
            npc.sendOk("You do not have enough mesos to change your emblem. Please come back when you have the required amount.")
        } else {
            npc.sendGuildEmblemEditor()
        }
    } else {
        npc.sendOk("Please come back when you wish to change your emblem.")
    }
} else {
    npc.sendOk("Please come back to me when you are a guild leader")
}