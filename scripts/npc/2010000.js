npc.sendBackNext("Hey, got a little bit of time? Well, my job is to collect items here and sell them elsewhere, but these days the monsters have become much more hostile so it's been difficult getting good items ... What do you think? Do you want to do some business with me?", false, true)

if (!npc.sendYesNo("The deal is simple. You get me something I need, I get you something you need. The problem is, I deal with a whole bunch of people, so the items I have to offer may change every time you see me. What do you think? Still want to do it?")) {
    npc.sendBackNext("Hmmm...it shouldn't be a bad deal for you. Come see me at the right time and you may get a much better item to be offered. Anyway, let me know when you have a change of heart.", true, true)
    return
}

var item = [4000073, 4000059, 4000076, 4000058, 4000078, 4000060, 4000062, 4000048, 4000081, 4000061, 4000070, 4000071, 4000072, 4000051, 4000055, 4000069, 4000052, 4000050, 4000057, 4000049, 4000056, 4000079, 4000053, 4000054, 4000080]

var chat = "Ok! First you need to choose the item that you'll trade with. The better the item, the more likely the chance that I'll give you something much nicer in return. #b"
for (var i = 0; i < item.length; i++) {
    chat += "\r\n#L" + i + "#100 #t" + item[i] + "#s#l"
}
npc.sendSelection(chat)
var select = npc.selection()

if (!npc.sendYesNo("Let's see, you want to trade your #b100 #t" + item[select] + "##k with my stuff, right? Before trading make sure you have an empty slot available on your use or etc. inventory. Now, do you really want to trade with me?")) {
    npc.sendBackNext("Hmmm ... it shouldn't be a bad deal for you at all. If you come at the right time I can hook you up with good items. Anyway if you feel like trading, feel free to come.", true, true)
    return
}

if (plr.itemQuantity(item[select]) < 100) {
    npc.sendOk("Hmmm... are you sure you have #b100 #t" + item[select] + "##k?")
    return
}

if (plr.getFreeSlots(2) < 1 || plr.getFreeSlots(4) < 1) {
    npc.sendBackNext("Your use and etc. inventory seem to be full. You need the free spaces to trade with me! Make room, and then find me...", true, true)
    return
}

var eQuestPrizes = []
eQuestPrizes[0] = [[2000001, 20], [2010004, 10], [2000003, 15], [4003001, 15], [2020001, 15], [2030000, 15]]
eQuestPrizes[1] = [[2000003, 20], [2000001, 30], [2010001, 40], [4003001, 20], [2040002, 1]]
eQuestPrizes[2] = [[2000002, 25], [2000006, 10], [2022000, 5], [4000030, 15], [2040902, 1]]
eQuestPrizes[3] = [[2000002, 30], [2000006, 15], [2020000, 20], [4003000, 5], [2041016, 1]]
eQuestPrizes[4] = [[2000002, 15], [2010004, 15], [2000003, 25], [4003001, 30], [2040302, 1]]
eQuestPrizes[5] = [[2000002, 30], [2000006, 15], [2020000, 20], [4003000, 5], [2040402, 1]]
eQuestPrizes[6] = [[2000002, 30], [2020000, 20], [2000006, 15], [4003000, 5], [2040402