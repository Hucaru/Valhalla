// Alcaster the Magician
var item = [2050003, 2050004, 4006000, 4006001];
var cost = [300, 400, 5000, 5000];

npc.sendSelection("What is it? \r\n#L0##bI want to buy something really rare.#l")

if (plr.level() < 30) {
    npc.sendOk("I am Alcaster the Magician. I have been studying all kinds of magic for over 300 years.")
} else {
    // Build selection string
    var chat = "Thanks to you, #bThe Book of Ancient#k is safely sealed. As a result, I used up about half of the power I have accumulated over the last 800 years...but can now die in peace. Would you happen to be looking for rare items by any chance? As a sign of appreciation for your hard work, I'll sell some items in my possession to you and ONLY you. Pick out the one you want! #b"
    for (var i = 0; i < item.length; i++) {
        chat += "\r\n#L" + i + "##t" + item[i] + "#(Price : " + cost[i] + " mesos)#l"
    }
    npc.sendSelection(chat)
    var select = npc.selection()

    var text1 = [
        "So the item you need is #bHoly Water#k, right? That's The item that cures the state of being sealed and cursed. It's not an easy item to get, but for you, I'll sell it for cheap. It'll cost you #b300 mesos#k per. How many would you like to buy?",
        "So the item you need is #bAll Cure Potion#k, right? That's The item that cures all. It's not an easy item to get, but for you, I'll sell it for cheap. It'll cost you #b400 mesos#k per. How many would you like to buy?",
        "So the item you need is #bThe Magic Rock#k, right? That's The item that possesses magical power and is used for high-quality skills. It's not an easy item to get, but for you, I'll sell it for cheap. It'll cost you #b5000 mesos#k per. How many would you like to buy?",
        "So the item you need is #bThe Summoning Rock#k, right? That's The item that possesses summoning power and is used for high-quality skills. It's not an easy item to get, but for you, I'll sell it for cheap. It'll cost you #b5000 mesos#k per. How many would you like to buy?"
    ]

    var num = npc.askNumber(text1[select], 1, 1, 100)

    if (npc.sendYesNo("Are you sure you want to buy #r" + num + " #t" + item[select] + "#(s)#k? It'll cost you " + cost[select] + " mesos per #t" + item[select] + "#, which will cost you #r" + (cost[select] * num) + "#k mesos total.")) {
        if (plr.mesos() < cost[select] * num || !plr.giveItem(item[select], num)) {
            npc.sendNext("Are you sure you have enough mesos? Please check if the Etc or Use windows of your Item Inventory is full, or if you have at least #r" + (cost[select] * num) + "#k mesos.")
        } else {
            plr.takeMesos(cost[select] * num)
            npc.sendNext("Thank you. If you need anything else, come see me anytime. I may have lost a lot of power, but I can still make magical items!")
        }
    } else {
        npc.sendNext("I see. Well, please understand that I carry many different items here. I'm only selling these items to you, so I won't be ripping you off in any way shape or form.")
    }
}

// Generate by kimi-k2-instruct