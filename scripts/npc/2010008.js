npc.sendSelection("Hi! My name is #bLea#k. I am in charge of the #bGuild Emblem#k. \r\n#L0##bI'd like to register a guild emblem.#l")
var sel = npc.selection()

if (sel == 0) {
    if (plr.guildRank() != 1) {
        npc.sendOk("You must be the Guild Leader to change the Emblem. Please tell your leader to speak with me.")
    } else {
        if (npc.sendYesNo("There is a fee of 500,000 mesos for creating a Guild Emblem. To further explain, a Guild Emblem is like a coat of arms that is unique to a guild, it will be displayed to the left of the guild name. How does that sound? Would you like to create a Guild Emblem?")) {
            plr.genericGuildMessage(18)
        }
    }
}

// Generate by kimi-k2-instruct