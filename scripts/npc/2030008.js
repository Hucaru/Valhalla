// Adobis - Door to Zakum

var minLevel = 50
var mapPartyPQ = 280010000
var mapJumpQuest = 280020000
var itemFireOre = 4031061
var itemBreathOfLava = 4031062
var itemGoldTooth = 4000082
var itemEyeOfFire = 4001017
var toothRequired = 30
var eyeReward = 5

if (plr.level() < minLevel) {
    npc.sendOk("You are not yet ready to face Zakum. Train a bit more and return when you are at least level " + minLevel + ".")
} else {
    var menuText =
        "The Door to Zakum lies ahead. How may I assist you?\r\n" +
        "#L0#Enter the Zakum Party Quest.#l\r\n" +
        "#L1#Enter the Zakum Jump Quest.#l\r\n" +
        "#L2#Exchange quest items for #t" + itemEyeOfFire + "#.#l"

    npc.sendSelection(menuText)
    var sel = npc.selection()

    if (sel == 0) {
        // Stage 1: Party Quest
        if (!plr.inParty()) {
            npc.sendOk("The Zakum Party Quest requires a party. Please form or join a party before attempting this quest.")
        } else if (!plr.isPartyLeader()) {
            npc.sendOk("Only your party leader can request entry for the party.\r\nPlease ask your leader to speak with me.")
        } else {
            var partySize = plr.partyMembersOnMapCount()
            if (partySize < 1) {
                npc.sendOk("You need at least one party member on this map to enter the Party Quest.")
            } else if (npc.sendYesNo("I can start the Zakum Party Quest for your party. You have " + partySize + " member(s) ready. Are you prepared?")) {
                // Start the party quest event
                plr.startPartyQuest("zakum_pq", 1)
            } else {
                npc.sendOk("Prepare well, and speak to me again when you are ready.")
            }
        }
    } else if (sel == 1) {
        if (npc.sendYesNo("I can send you to the Zakum Jump Quest (#m" + mapJumpQuest + "#).\r\nWould you like to go now?")) {
            plr.warp(mapJumpQuest)
        } else {
            npc.sendOk("Very well. Let me know when you wish to attempt it.")
        }
    } else if (sel == 2) {
        // Exchange items -> Eyes of Fire
        var haveOre = plr.itemCount(itemFireOre)
        var haveBreath = plr.itemCount(itemBreathOfLava)
        var haveTooth = plr.itemCount(itemGoldTooth)

        var reqText =
            "To receive (" + eyeReward + ") #t" + itemEyeOfFire + "#, I require:\r\n" +
            "- (1) #t" + itemFireOre + "#\r\n" +
            "- (1) #t" + itemBreathOfLava + "#\r\n" +
            "- (" + toothRequired + ") #t" + itemGoldTooth + "#\r\n\r\n" +
            "Your inventory:\r\n" +
            "- #t" + itemFireOre + "#: (" + haveOre + ")\r\n" +
            "- #t" + itemBreathOfLava + "#: (" + haveBreath + ")\r\n" +
            "- #t" + itemGoldTooth + "#: (" + haveTooth + ")\r\n\r\n" +
            "Proceed with the exchange?"

        if (haveOre < 1 || haveBreath < 1 || haveTooth < toothRequired) {
            npc.sendOk(
                "You do not have the required items.\r\n" +
                "Bring me:\r\n" +
                "- (1) #t" + itemFireOre + "#\r\n" +
                "- (1) #t" + itemBreathOfLava + "#\r\n" +
                "- (" + toothRequired + ") #t" + itemGoldTooth + "#."
            )
        } else if (npc.sendYesNo(reqText)) {
            if (!plr.giveItem(itemEyeOfFire, eyeReward)) {
                npc.sendOk("Please make sure you have enough space in your Etc. inventory, then try again.")
            } else {
                var ok1 = plr.removeItemsByID(itemFireOre, 1)
                var ok2 = plr.removeItemsByID(itemBreathOfLava, 1)
                var ok3 = plr.removeItemsByID(itemGoldTooth, toothRequired)

                if (!(ok1 && ok2 && ok3)) {
                    plr.removeItemsByID(itemEyeOfFire, eyeReward)
                    npc.sendOk("An error occurred while taking the required items.\r\nPlease ensure the items are tradable and try again.")
                } else {
                    npc.sendOk("Here are your (" + eyeReward + ") #t" + itemEyeOfFire + "#.\r\nGood luck challenging Zakum.")
                }
            }
        } else {
            npc.sendOk("Come back when you are ready to exchange.")
        }
    }
}