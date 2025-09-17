npc.sendNext("Let's see, tongue of mole and beak of owl in proper... Ah, blast it! I nearly forgot the binding powder. That could have been... Oh. How long have you been standing there? Speak up next time! I get very absorbed in my work.")

var firstPick = npc.sendMenu(
    "I am Eurek the Alchemist. There is still much for me to learn, but I am sure I can help you. What do you need?",
    "#bMake The Magic Rock",
    "#bMake The Summoning Rock"
)

// firstPick returns 0 for Magic, 1 for Summoning
var items, matSet, matSetQty, costSet, itemid
if (firstPick === 0) {
    items     = [4006000, 4006000, 4006000, 4006000, 4006000, 4006000]
    matSet    = [[4000046, 4000027, 4021001],
                 [4000025, 4000049, 4021006],
                 [4000129, 4000130, 4021002],
                 [4000074, 4000057, 4021005],
                 [4000054, 4000053, 4021003],
                 [4000238, 4000241, 4021000]]
    matSetQty = [[20, 20, 1],
                 [20, 20, 1],
                 [15, 15, 1],
                 [15, 15, 1],
                 [7,  7,  1],
                 [15, 15, 1]]
    costSet   = [4000, 4000, 4000, 4000, 4000, 4000]
    itemid    = "The Magic Rocks"
} else {
    items     = [4006001, 4006001, 4006001, 4006001, 4006001, 4006001]
    matSet    = [[4000028, 4000027, 4011001],
                 [4000014, 4000056, 4011003],
                 [4000132, 4000128, 4011005],
                 [4000074, 4000069, 4011002],
                 [4000080, 4000079, 4011004],
                 [4000226, 4000237, 4011001]]
    matSetQty = [[20, 20, 1],
                 [20, 20, 1],
                 [15, 15, 1],
                 [15, 15, 1],
                 [7,  7,  1],
                 [15, 15, 1]]
    costSet   = [4000, 4000, 4000, 4000, 4000, 4000]
    itemid    = "The Summoning Rocks"
}

var chat = ""
if (firstPick === 0) {
    chat = "Ha. #bThe Magic Rock#k is a mysterious rock that only I can make. Many Explorers need it when they use very powerful skills. There are six ways to make The Magic Rock. Which do you want to use? #b"
} else {
    chat = "Ha. #bThe Summoning Rock#k is a mysterious rock that only I can make. Many Explorers need it when they use very powerful skills. There are six ways to make the The Summoning Rock. Which method do you want to use? #b"
}
for (var i = 0; i < 6; i++) {
    chat += "\r\n#L" + i + "#make it using #z" + matSet[i][0] + "# and #z" + matSet[i][1] + "##l"
}

npc.sendSelection(chat)
var secondPick = npc.selection()

var item   = items[secondPick]
var mat    = matSet[secondPick]
var matQty = matSetQty[secondPick]
var cost   = costSet[secondPick]

chat = "To make #b5 " + itemid + "#k, I'll need the following items. Most of them can be obtained through hunting, so it won't be terribly difficult for you to get them. What do you think? Do you want some? \r\n#b"
for (i = 0; i < mat.length; i++) {
    chat += "\r\n#v" + mat[i] + "# #t" + mat[i] + "# x " + matQty[i]
}
chat += "\r\n#v4031138#" + cost + " Mesos"

if (npc.sendYesNo(chat)) {
    // check supplies
    var missing = false
    for (i = 0; i < mat.length; i++) {
        if (plr.itemCount(mat[i]) < matQty[i]) {
            missing = true
            break
        }
    }
    if (missing || plr.mesos() < cost) {
        npc.sendNext("Please check and see if you have all the items needed, or if your etc. inventory is full or not.")
    } else {
        // take items & mesos
        for (i = 0; i < mat.length; i++) {
            plr.removeItemsByID(mat[i], matQty[i])
        }
        plr.giveMesos(-cost)
        plr.giveItem(item, 5)
        npc.sendNext("Here, take the 5 pieces of #b" + itemid + "#k. Even I have to admit, this is a masterpiece. Alright, if you need my help down the road, by all means come back and talk to me!")
    }
} else {
    npc.sendNext("Not enough materials, huh? No worries. Just come see me once you gather up the necessary items. There are numerous ways to obtain them, whether it be hunting or purchasing it from others, so keep going.")
}