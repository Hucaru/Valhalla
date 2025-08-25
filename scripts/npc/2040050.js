npc.sendBackNext("Let's see, tongue of mole and beak of owl in proper... Ah, blast it! I nearly forgot the binding powder. That could have been... Oh. How long have you been standing there? Speak up next time! I get very absorbed in my work.", false, true)

var menu = "I am Eurek the Alchemist. There is still much for me to learn, but I am sure I can help you. What do you need? \r\n#L0##bMake The Magic Rock#l\r\n#L1#Make The Summoning Rock#l"
npc.sendSelection(menu)
var select = npc.selection()

var items, itemid, item, mat, matQty, cost

if (select == 0) {
    items = [[4000046, 4000027], [4000025, 4000049], [4000129, 4000130], [4000074, 4000057], [4000054, 4000053], [4000238, 4000241]]
    var chat = "Ha. #bThe Magic Rock#k is a mysterious rock that only I can make. Many Explorers need it when they use very powerful skills. There are six ways to make The Magic Rock. Which do you want to use? #b"
    for (var i = 0; i < items.length; i++)
        chat += "\r\n#L" + i + "#make it using #z" + items[i][0] + "# and #z" + items[i][1] + "##l"
    npc.sendSelection(chat)
} else if (select == 1) {
    items = [[4000028, 4000027], [4000014, 4000056], [4000132, 4000128], [4000074, 4000069], [4000080, 4000079], [4000226, 4000237]]
    var chat = "Ha. #bThe Summoning Rock#k is a mysterious rock that only I can make. Many Explorers need it when they use very powerful skills. There are six ways to make the The Summoning Rock. Which method do you want to use? #b"
    for (var i = 0; i < items.length; i++)
        chat += "\r\n#L" + i + "#make it using #z" + items[i][0] + "# and #z" + items[i][1] + "##l"
    npc.sendSelection(chat)
}

var selectItem = npc.selection()

if (select == 0) {
    items = [4006000, 4006000, 4006000, 4006000, 4006000]
    var matSet = [[4000046, 4000027, 4021001], [4000025, 4000049, 4021006], [4000129, 4000130, 4021002], [4000074, 4000057, 4021005], [4000054, 4000053, 4021003], [4000238, 4000241, 4021000]]
    var matSetQty = [[20, 20, 1], [20, 20, 1], [15, 15, 1], [15, 15, 1], [7, 7, 1], [15, 15, 1]]
    var costSet = [4000, 4000, 4000, 4000, 4000]
    itemid = "The Magic Rocks"
} else if (select == 1) {
    items = [4006001, 4006001, 4006001, 4006001, 4006001]
    var matSet = [[4000028, 4000027, 4011001], [4000014, 4000056, 4011003], [4000132, 4000128, 4011005], [4000074, 4000069, 4011002], [4000080, 4000079, 4011004], [4000226, 4000237, 4011001]]
    var matSetQty = [[20, 20, 1], [20, 20, 1], [15, 15, 1], [15, 15, 1], [7, 7, 1], [15, 15, 1]]
    var costSet = [4000, 4000, 4000, 4000, 4000]
    itemid = "The Summoning Rocks"
}

item =