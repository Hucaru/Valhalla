// Cathy – Mini-game NPC (stateless)
// Maps: 100000203 (Bowman village game center) or any copy of it

// ------------------------------------------------------------------
// 1. Top greeting
// ------------------------------------------------------------------
npc.sendSelection(
    "Hey! Looks like you could use a breather. You should enjoy life, like I do! Well, if you have the right items, you can trade them to me for an item you can use to play mini-games. So what'll it be?\r\n"
    + "#L0##bCreate a mini-game item#l\r\n"
    + "#L1#Listen to the explanation on mini-games#l"
)
var mainSelect = npc.selection()

// ------------------------------------------------------------------
// 2. Branch on main choice
// ------------------------------------------------------------------
if (mainSelect === 0) {

    // ------------------------------
    // CREATE mini-game items
    // ------------------------------
    npc.sendSelection(
        "Would you like to create a mini-game item? They'll allow you to start up a mini-game nearly anywhere.\r\n"
        + "#L6##bOmok Set#l\r\n"
        + "#L7#Memory Set#l"
    )
    var gameType = npc.selection()

    if (gameType === 6) {

        // -------- Omok section --------
        npc.sendBackNext("Looks like you want to play #bOmok#k. You'll need an Omok Set to play. Only those with that item can join a room for playing Omok. You can open an Omok room nearly anywhere, except for the Free Market and waiting maps.", false, true)
        npc.sendBackNext("There are various Omok Sets with different visual aesthetics.", true, true)

        // List Omok Sets (6 choices, 0-based)
        var omok    = [4080000, 4080001, 4080002, 4080003, 4080004, 4080005];
        var omok1   = [4030000, 4030000, 4030000, 4030010, 4030011, 4030011];
        var omok2   = [4030001, 4030010, 4030011, 4030001, 4030010, 4030001];
        var omok3   = ["Slime", "Slime", "Slime", "Octopus", "Pig", "Pig"];
        var omok4   = ["Mushroom", "Octopus", "Pig", "Mushroom", "Octopus", "Mushroom"];

        var omokSel = npc.sendMenu(
            "Which Omok Set do you want to make?#b",
            "#z" + omok[0] + "#",
            "#z" + omok[1] + "#",
            "#z" + omok[2] + "#",
            "#z" + omok[3] + "#",
            "#z" + omok[4] + "#",
            "#z" + omok[5] + "#"
        )

        npc.sendBackNext(
            "You want #b#t" + omok[omokSel] + "##k? Hmm... I can make it for you if I had the right materials. If you can bring me #rOmok Piece : " + omok3[omokSel] + ", Omok Piece : " + omok4[omokSel] + ", and Omok Table#k items. I'm pretty sure monsters drop those...",
            true, true
        )

        // Check Omok items
        if (plr.itemCount(omok1[omokSel]) < 1 || plr.itemCount(omok2[omokSel]) < 1 || plr.itemCount(4030009) < 1) {
            npc.sendOk("#rYou don't have all the required pieces.#k")
        } else if (npc.sendYesNo("Wait right there and I'll get to work.")) {
            var etc = plr.getInventory(4); // 4 = etc inventory
            if (etc.getNumFreeSlot() < 1) {
                npc.sendOk("#rNot enough free slots in Etc tab.#k")
            } else {
                plr.removeItemsByID(omok1[omokSel], 1)
                plr.removeItemsByID(omok2[omokSel], 1)
                plr.removeItemsByID(4030009, 1)
                plr.giveItem(omok[omokSel], 1)
                npc.sendBackNext(
                    "Here's your #b#t" + omok[omokSel] + "##k! Now you can set up an Omok room whenever you want.",
                    true, true
                )
                npc.sendBackNext("I'll be here a while, so let me know if you have any questions about Omok.", true, true)
            }
        }

    } else if (gameType === 7) {

        // -------- Memory section --------
        npc.sendBackNext("You want to make #bA set of Memory Cards#k? Hmm... I'll need several items to make A set of Memory Cards. You can get Monster Card by defeating monsters here and there on the island. Get me 15 of those and I'll make you A set of Memory Cards.", false, true)

        if (plr.itemCount(4030012) < 15) {
            npc.sendOk("#rYou don't have 15 Monster Cards.#k")
        } else if (npc.sendYesNo("All right! Time to show off my skills. Wait a moment and I'll make you #rA set of Memory Cards#k.")) {
            var etc = plr.getInventory(4)
            if (etc.getNumFreeSlot() < 1) {
                npc.sendOk("#rNot enough free slots in Etc tab.#k")
            } else {
                plr.removeItemsByID(4030012, 15)
                plr.giveItem(4080100, 1)
                npc.sendBackNext("Here you are, #bA set of Memory Cards#k. Use the item to start a match and play with others!", true, true)
                npc.sendBackNext("I'll be here a while, so let me know if you have any questions about Memory.", true, true)
            }
        }
    }

} else if (mainSelect === 1) {

    // ------------------------------
    // EXPLAIN mini-games
    // ------------------------------
    npc.sendSelection(
        "Do you want to learn about mini-games? Fine! Ask me anything. Which mini-game should I explain to you?\r\n"
        + "#L8##bOmok#l\r\n"
        + "#L9#Memory#l"
    )
    var explain = npc.selection()

    if (explain === 8) {

        // -------- Omok explanation --------
        npc.sendBackNext("Omok is played by placing tiles known as 'stones' on the board. The first player to get five stones in a row vertically, horizontally, or diagonally wins. Also, only someone with an #bOmok Set#k can create a game, though anyone can join. Omok can be played just about anywhere.", false, true)
        npc.sendBackNext("You need #r100 mesos#k for every game of Omok...", true, true)
        npc.sendBackNext("After you've entered a game, press the #bReady#k button...", true, true)
        npc.sendBackNext("When the room is first opened and the first game starts, the #bhost goes first#k...", true, true)
        npc.sendBackNext("If you make a mistake, you can request a #btake back#k...", true, true)
        npc.sendBackNext("After the first match, the loser of each previous map will get to go first...", true, true)
        npc.sendBackNext("I wonder if there's anyone who can surpass me? That's it for Omok. Come to me again if you have questions later.", true, true)

    } else if (explain === 9) {

        // -------- Memory explanation --------
        npc.sendBackNext("You don't know how to play Memory? That's okay. Basically, you need find the matching pairs of face-down monster cards...", false, true)
        npc.sendBackNext("You'll need #r100 mesos#k for every game of Memory...", true, true)
        npc.sendBackNext("After you've entered a game, press the #bReady#k button...", true, true)
        npc.sendBackNext("One more thing... You'll need to decide how many cards you want to use...", true, true)
        npc.sendBackNext("At the start of the first game, the #bhost gets the first move#k...", true, true)
        npc.sendBackNext("If both you and your opponent have the same number of matches...", true, true)
        npc.sendBackNext("After the first match, the loser of each previous map will get to go first...", true, true)
        npc.sendBackNext("I wonder if there's anyone who can surpass me? That's it for Memory. Come to me again if you have questions later.", true, true)
    }
}