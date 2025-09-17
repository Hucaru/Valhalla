// Hello! I'm #b#p2040014##k...
var sel0 = npc.sendSelection("Hello! I'm #b#p2040014##k, and l'm in charge of mini-games around these parts. You look like you could use a good mini-game, and I can help! All right... What do you want to do? \r\n#L0##bCreate a mini-game item#l\r\n#L1#Listen to the explanation on mini-games#l")

if (sel0 === 0) {
    // Create mini-game item
    var sel1 = npc.sendSelection("Would you like to create a mini-game item? They'll allow you to start up a min-game nearly anywhere. \r\n#L0##bOmok Set#l\r\n#L1#Memory Set#l")
    
    if (sel1 === 0) {
        // Omok Set
        npc.sendBackNext("Looks like you want to play #bOmok#k. You'll need an Omok Set to play. Only those with that item can join a room for playing Omok. You can open an Omok room nearly anywhere, except for the Free Market and waiting maps.", false, true)
        
        var omok      = [4080006, 4080007, 4080008, 4080009, 4080010, 4080011]
        var omok1     = [4030013, 4030013, 4030014, 4030015, 4030015, 4030015]
        var omok2     = [4030014, 4030016, 4030016, 4030013, 4030014, 4030016]
        var omok3     = ["Bloctopus", "Bloctopus", "Pink Teddy", "Panda Teddy", "Panda Teddy", "Panda Teddy"]
        var omok4     = ["Pink Teddy", "Trixter", "Trixter", "Bloctopus", "Pink Teddy", "Trixter"]
        
        var chat = "There are various Omok Sets with different visual aesthetics. Which Omok Set do you want to make? #b"
        for (var i = 0; i < omok.length; i++) chat += "\r\n#L"+i+"##z" + omok[i] + "##l"
        
        var selType = npc.sendSelection(chat)
        
        npc.sendBackNext("You want #b#t" + omok[selType] + "##k? Hmm... I can make it for you if I had the right materials. If you can bring me #rOmok Piece : " + omok3[selType] + ", Omok Piece : " + omok4[selType] + ", and Omok Table#k items. I'm pretty sure monsters drop those...", true, true)
        
        var hasPieces = (plr.itemCount(omok1[selType]) >= 1) &&
                        (plr.itemCount(omok2[selType]) >= 1) &&
                        (plr.itemCount(4030009) >= 1)
                        
        if (!hasPieces) {
            npc.sendOk("Are you sure you have #bOmok Piece : " + omok3[selType] + ", Omok Piece : " + omok4[selType] + ", and Omok Table#k?")
        } else {
            npc.sendBackNext("Wow! Aren't these the #rOmok Piece : " + omok3[selType] + ", Omok Piece : " + omok4[selType] + ", and Omok Table#k items? That should be everything I need. Wait right there and l'll get to work.", true, true)
            
            if (plr.giveItem(omok[selType], 1)) {
                plr.takeItem(omok1[selType], 0, 1)  // assume slot 0 or engine handles
                plr.takeItem(omok2[selType], 0, 1)
                plr.removeItemsByID(4030009, 1)
                
                npc.sendBackNext("Here's your #b#t" + omok[selType] + "##k! Now you can set up an Omok room whenever you want. Just use your Omok set to open a room and play with others. Something good might even happen if your record is impressive enough. I'll be cheering you on from here, so do your best.", true, true)
                npc.sendBackNext("I'll be here a while, so let me know if you have any questions about Omok. Work hard to become the best mini-gamer around! Maybe one day you'll surpass me. As if. Anywho, seeya.", true, true)
            } else {
                npc.sendOk("Not enough space in your Etc tab!")
            }
        }
        
    } else {
        // Memory Set
        npc.sendBackNext("You want to make #bA set of Memory Cards#k? Hmm... I'll need several items to make A set of Memory Cards. You can get Monster Card by defeating monsters here and there on the island. Get me 15 of those and I'll make you A set of Memory Cards.", false, true)
        
        if (plr.itemCount(4030012) < 15) {
            npc.sendOk("You don't have 15 Monster Cards!")
        } else {
            npc.sendBackNext("Wow! That's definitely 15 #rMonster Card#k items. It must've been hard getting them all. All right! Time to show off my skills. Wait a moment and I'll make you #rA set of Memory Cards#k.", true, true)
            
            if (plr.giveItem(4080100, 1)) {
                plr.removeItemsByID(4030012, 15)
                
                npc.sendBackNext("Here you are, #bA set of Memory Cards#k. Use the item to start a match and play with others! You can open a Memory room nearly anywhere, except for the Free Market and waiting maps. Something good might happen if you get a nice record. I'll be cheering you on from here, so do your best.", true, true)
                npc.sendBackNext("I'll be here a while, so let me know if you have any questions about Memory. Work hard to become the best mini-gamer around! Maybe one day you'll surpass me. As if. Anywho, seeya.", true, true)
            } else {
                npc.sendOk("Not enough space in your Etc tab!")
            }
        }
    }
    
} else {
    // Listen to explanations
    var sel2 = npc.sendSelection("Do you want to learn about mini-games? Fine! Ask me anything. Which mini-game should I explain to you? \r\n#L0##bOmok#l\r\n#L1#Memory#l")
    
    if (sel2 === 0) {
        // Omok explanation chain
        npc.sendBackNext("Omok is played by placing tiles known as 'stones' on the board. The first player to get five stones in a row vertically, horizontally, or diagonally wins. Also, only someone with an #bOmok Set#k can create a game, though anyone can join. Omok can be played just about anywhere.", false, true)
        npc.sendBackNext("You need #r100 mesos#k for every game of Omok. Even if you don't have an Omok Set, you can join someone else with an open game. That's assuming you've got 100 mesos. You'll be kicked out of the room if you run out of mesos.", true, true)
        npc.sendBackNext("After you've entered a game, press the #bReady#k button to show you're ready to start. The host can then press the #bStart#k button to start the game. If someone you don't want to play with joins a game you're hosting, you can kick them out by pressing the 'X' on the top right of the window.", true, true)
        npc.sendBackNext("When the room is first opened and the first game starts, the #bhost goes first#k. Make sure to make your move within the time limit, or your opponent will get to go again. Note that you can't place tiles 3 x 3, unless placing your tile elsewhere would end the game. That means placing tiles 3 x 3 is only allowed if it's necessary for defense! Another thing! Only 5 in a row counts as Omok! That means you can't win by connecting #r6 or 7#k.", true, true)
        npc.sendBackNext("If you make a mistake, you can request a #btake back#k. If your opponent accepts, you can withdraw the stone you just placed and set it somewhere else. If something comes up and you have to leave, you can request a #bdraw#k. If your opponent accepts, the game ends in a draw. Requesting a draw is always better than destroying a friendship.", true, true)
        npc.sendBackNext("After the first match, the loser of each previous map will get to go first. Also, there's no wandering off in the middle of a game. If you really want to leave, you need to ask to #bdrop out or request a draw#k. Remember that Giving up counts as an automatic loss. Otherwise, when you press Exit, you'll leave the room after the completion of the current match.", true, true)
        npc.sendBackNext("I wonder if there's anyone who can surpass me? That's it for Omok. Come to me again if you have questions later.", true, true)
        
    } else {
        // Memory explanation chain
        npc.sendBackNext("You don't know how to play Memory? That's okay. Basically, you need find the matching pairs of face-down monster cards. The person with the most matches once all cards are face up is the winner. One player must have #bA set of Memory Cards#k to host the game, and it can be played anywhere.", false, true)
        npc.sendBackNext("You'll need #r100 mesos#k for every game of Memory. Even if you don't have #bA set of Memory Cards#k, you can enter a join someon with an open game. However, you won't be able to enter a room without 100 mesos, and you'll be forced to exit if you run out of mesos.", true, true)
        npc.sendBackNext("After you've entered a game, press the #bReady#k button to show you're ready to start. The host can then press the #bStart#k button to start the game. If someone you don't want to play with joins a game you're hosting, you can kick them out by pressing the 'X' on the top right of the window.", true, true)
        npc.sendBackNext("One more thing... You'll need to decide how many cards you want to use when opening a game of Memory. There are 3 modes: 3x4, 4x5, and 5x6 which result in a playing field of 12, 20, or 30 cards respectively. If you want to change the board size, you'll need to make an new room.", true, true)
        npc.sendBackNext("At the start of the first game, the #bhost gets the first move#k. Be sure to flip a card within the time limit or your opponent will get to go again. You get an extra move if you match a pair of cards on your turn. Remembering where cards appear is the key to winning.", true, true)
        npc.sendBackNext("If both you and your opponent have the same number of matches, the one who got the most cards right in a row wins. If you suddenly need to use the restroom or you know that you're going to lose, you can request a #bdraw#k. If your opponent accepts, the game ends in a draw. I would recommend a draw if you want to avoid destroying friendships. Isn't that better than fighting?", true, true)
        npc.sendBackNext("After the first match, the loser of each previous map will get to go first. Also, there's no wandering off in the middle of a game. If you really want to leave, you need to ask to #bdrop out or request a draw#k. Remember that Giving up counts as an automatic loss. Otherwise, when you press Exit, you'll leave the room after the completion of the current match.", true, true)
        npc.sendBackNext("I wonder if there's anyone who can surpass me? That's it for Memory. Come to me again if you have questions later.", true, true)
    }
}