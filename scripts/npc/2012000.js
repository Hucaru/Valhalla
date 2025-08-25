// Ticket Agent – Orbis Station
var menu = "I can guide you to the right ship to reach your destination. Where are you headed? \r\n#L0##bVictoria Island#l\r\n#L1#Ludibrium Castle#l\r\n#L2#Leafre#l\r\n#L3#Mu Lung#l\r\n#L4#Ariant#l\r\n#L5#Ereve#l\r\n#L6#Edelstein#l"
npc.sendSelection(menu)
var sel = npc.selection()

switch (sel) {
    case 0: // Victoria Island
        npc.sendBackNext("You're headed to Victoria Island? Oh, it's a beautiful island with a variety of villages. A #b1-seater ship for Victoria is always standing by#k for you to use.", false, true)
        npc.sendBackNext("Talk to #bIsa the Platform Guide#k on the right if you would like to take the Airship to Victoria. If anyone can show you the way, it's Isa.", true, true)
        break
    case 1: // Ludibrium Castle
        npc.sendBackNext("You're headed to Ludibrium Castle at Ludus Lake? It's such a fun village made of toys. A #b1-seater ship for Ludibrium Fortress is always standing by#k for you to use.", false, true)
        npc.sendBackNext("Talk to #bIsa the Platform Guide#k on the right if you would like to take the Airship to Ludibrium. If anyone can show you the way, it's Isa.", true, true)
        break
    case 2: // Leafre
        npc.sendBackNext("You're headed to Leafre in Minar Forest? I love that quaint little village of Halflingers. A #b1-seater ship for Leafre is always standing by#k for you to use.", false, true)
        npc.sendBackNext("Talk to #bIsa the Platform Guide#k on the right if you would like to take the Airship to Leafre. If anyone can show you the way, it's Isa.", true, true)
        break
    case 3: // Mu Lung
        npc.sendBackNext("Are you heading towards Mu Lung in the Mu Lung Temple? I'm sorry, but there's no ship that flies from Orbis to Mu Lung. There is another way to get there, though. There's a #bCrane that runs a cab service for 1 that's always available#k, so you'll get there as soon as you wish.", false, true)
        npc.sendBackNext("Unlike the ships that fly for free, however, this cab requires a set fee. This personalized flight to Mu Lung will cost you #b1,500 mesos#k, so please have the fee ready before riding the Crane.", true, true)
        npc.sendBackNext("Talk to #bIsa the Platform Guide#k on the right if you would like to take the Crane to Mu Lung. If anyone can show you the way, it's Isa.", true, true)
        break
    case 4: // Ariant
        npc.sendBackNext("You're headed to Ariant in the Nihal Desert? The people living there have a passion as hot as the desert. A #bship that heads to Ariant is always standing by#k for you to use.", false, true)
        npc.sendBackNext("Talk to #bIsa the Platform Guide#k on the right if you would like to take the Genie to Ariant. If anyone can show you the way, it's Isa.", true, true)
        break
    case 5: // Ereve
        npc.sendBackNext("Are you heading toward Ereve? It's a beautiful island blessed with the presence of the Shinsoo the Holy Beast and Empress Cygnus. #bThe boat is for 1 person and it's always readily available#k so you can travel to Ereve fast.", false, true)
        npc.sendBackNext("Talk to #bIsa the Platform Guide#k on the right if you would like to take the ship to Ereve. If anyone can show you the way, it's Isa.", true, true)
        break
    case 6: // Edelstein
        npc.sendBackNext("Are you going to Edelstein? The brave people who live there constantly fight the influence of dangerous monsters. #b1-person Airship to Edelstein is always on standby#k, so you can use it at any time.", false, true)
        npc.sendBackNext("Talk to #bIsa the Platform Guide#k on the right if you would like to take the ship to Edelstein. If anyone can show you the way, it's Isa.", true, true)
        break
}

