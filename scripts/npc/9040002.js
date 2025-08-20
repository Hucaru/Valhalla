npc.sendSelection("We, the Union of Guilds, have been trying to decipher 'Emerald Tablet,' a treasured old relic, for a long time. As a result, we have found out that Sharenian, the mysterious country from the past, lay asleep here. We also found out that clues of Rubian, a legendary, mythical jewelry, may be here at the remains of Sharenian. This is why the Union of Guilds have opened Guild Quest to ultimately find Rubian.\r\n#L0##bWhat's Sharenian?#l\r\n#L1#Rubian?#l\r\n#L2#Guild Quest?#l\r\n#L3#I'm fine now.#l")
var select = npc.selection()

if (select == 0) {
    npc.sendBackNext("Sharenian was a literate civilization from the past that had control over every area of the Victoria Island. The Temple of Golem, the Shrine in the deep part of the Dungeon, and other old architectural constructions where no one knows who built it are indeed made during the Sharenian times.", false, true)
    npc.sendBackNext("The last king of Sharenian was a gentleman named Sharen Ill, and apparently he was a very wise and compassionate king. But one day, the whole kingdom collapsed, and there was no explanation made for it.", true, false)
} else if (select == 1) {
    npc.sendBackNext("Rubian is a legendary jewel that brings eternal youth to the one that possesses it. Ironically, it seems like everyone that had Rubian ended up downtrodden, which should explain the downfall of Sharenian.", false, false)
} else if (select == 2) {
    npc.sendBackNext("I've sent groups of the explorers to Sharenian before, but none of them ever came back, which prompted us to start the Guild Quest. We've been waiting for guilds that are strong enough to take on tough challenges, guilds like yours.", false, true)
    npc.sendBackNext("The ultimate goal of this Guild Quest is to explore Sharenian and find Rubian. This is not a task where power solves everything. Teamwork is more important here.", true, false)
} else if (select == 3) {
    npc.sendOk("Really? If you have anything else to ask, please feel free to talk to me.")
}

// Generate by kimi-k2-instruct