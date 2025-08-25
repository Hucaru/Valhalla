npc.sendSelection("Hey! Looks like you could use a breather. You should enjoy life, like I do! Well, if you have the right items, you can trade them to me for an item you can use to play mini-games. So what'll it be? \r\n#L0##bCreate a mini-game item#l\r\n#L1#Listen to the explanation on mini-games#l")
var sel0 = npc.selection()

if (sel0 == 0) {
    npc.sendSelection("Would you like to create a mini-game item? They'll allow you to start up a min-game nearly anywhere. \r\n#L0##bOmok Set#l\r\n#L1#Memory Set#l")
    var sel1 = npc.selection()

    if (sel1 == 0) {
        npc.sendBackNext("Looks like you want to play #bOmok#k. You'll need an Omok Set to play. Only those with that item can join a room for playing Omok. You can open an Omok room nearly anywhere, except for the Free Market and waiting maps.", false, true)

        var omok = [4080000, 4080001, 4080002, 4080003, 4080004, 4080005]
        var omok1 = [4030000, 4030000, 4030000, 4030010, 4030011, 4030011]
        var omok2 = [4030001, 4030010, 4030011, 4030001, 4030010, 4030001]
        var omok3 = ["Slime", "Slime", "Slime", "Octopus", "Pig", "Pig"]
        var omok4 = ["Mushroom", "Octopus", "Pig", "Mushroom", "Octopus", "Mushroom"]

        var chat = "There are various Omok Sets with different visual aesthetics. Which Omok Set do you want to make? #b"
        for (var i = 0; i < omok.length; i++) {
            chat += "\r\n#L" + i + "##z" + omok[i] + "##l"
        }
        npc.sendSelection(chat)
        var select = npc.selection()

        npc.sendBackNext("You want #b#t" + omok[select] + "##k? Hmm... I can make it for you if I had the right materials. If you can bring me #rOmok Piece : " + omok3[select] + ", Omok Piece : " + omok4[select] + ", and Omok Table#k items. I'm pretty sure monsters drop those...", true, true)

        if (plr.itemQuantity(omok1[select]) < 1 || plr.itemQuantity(omok2[select]) < 1 || plr.itemQuantity(4030009) < 1) {
            npc.sendBackNext("Are you sure you have #bOmok Piece : " + omok3[select] + ", Omok Piece : " + omok4[select] + ", and Omok Table#k? Maybe you just don't have enough space in your Etc tab.", true, true)
        } else {
            if (plr.getInventory(4).getNumFreeSlot() < 1) {
                npc.sendBackNext("Are you sure you have #bOmok Piece : " + omok3[select] + ", Omok Piece : " + omok4[select] + ", and Omok Table#k? Maybe you just don't have enough space in your Etc tab.", true, true)
            } else {
                plr.takeItem(omok1[select], 1)
                plr.takeItem(omok2[select], 1)
                plr.takeItem(4030009, 1)
                plr.giveItem(omok[select], 1)
                npc.sendBackNext("Here's your #b#t" + omok[select] + "##k! Now you can set up an Omok room whenever you want. Just use your Omok set to open a room and play with others. Something good might even happen if your record is impressive enough. I'll be cheering you on from here, so do your best.", true, true)
                npc.sendBackNext("I'll be here a while, so let me know if you have any questions about Omok. Work hard to become the best mini-gamer around! Maybe one day you'll surpass me. As if. Anywho, seeya.", true, true)
            }
        }
    } else if (sel1 == 1) {
        npc