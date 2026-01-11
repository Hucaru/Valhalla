// Aura

var mapReturn = 211042300
var itemKeys = 4001016            // Keys
var keysRequired = 7
var itemFireOre = 4031061         // Fire Ore (warps after exchange)

var itemDocs = 4001015            // Documents
var docsRequired = 32
var itemReturnScroll = 2030007    // Dead Mine Scroll
var scrollReward = 5              // (Optional) no warp

var questStage1 = 7000            // Quest ID for Stage 1 completion tracking

var menu =
    "What would you like to do?\r\n" +
    "#L0#Leave this place.#l\r\n" +
    "#L1#Exchange keys for #t" + itemFireOre + "# and warp back to #m"+ mapReturn +"# . #l\r\n" +
    "#L2#(Optional) Exchange documents for #t" + itemReturnScroll + "# #l"

npc.sendSelection(menu)
var sel = npc.selection()

if (sel == 0) {
    if (npc.sendYesNo("Do you want to return to #m" + mapReturn + "# now?")) {
        plr.warp(mapReturn)
    } else {
        npc.sendOk("Alright. Let me know if you change your mind.")
    }
} else if (sel == 1) {
    // Keys -> Fire Ore, then warp
    var haveKeys = plr.itemCount(itemKeys)
    if (haveKeys < keysRequired) {
        npc.sendOk(
            "You do not have enough #t" + itemKeys + "# .\r\n" +
            "Required: (" + keysRequired + "), You have: (" + haveKeys + ")"
        )
    } else if (npc.sendYesNo(
        "Exchange (" + keysRequired + ") #t" + itemKeys + "# for (1) #t" + itemFireOre + "#\r\n" +
        "and then return to #m" + mapReturn + "#?"
    )) {
        if (!plr.giveItem(itemFireOre, 1)) {
            npc.sendOk("Please make sure you have enough inventory space, then try again.")
        } else {
            var took = plr.removeItemsByID(itemKeys, keysRequired)
            if (!took) {
                plr.removeItemsByID(itemFireOre, 1) // rollback
                npc.sendOk("An error occurred while taking the keys. Please try again.")
            } else {
                // Mark Stage 1 as completed
                plr.setQuestData(questStage1, "end")
                npc.sendOk("Exchange complete. You have completed Stage 1! Returning you now.")
                plr.warp(mapReturn)
            }
        }
    } else {
        npc.sendOk("Alright. Come back when you are ready.")
    }
} else if (sel == 2) {
    // Optional: Documents -> Dead Mine Scrolls
    var haveDocs = plr.itemCount(itemDocs)
    if (haveDocs < docsRequired) {
        npc.sendOk(
            "You do not have enough #t" + itemDocs + "#.\r\n" +
            "Required: (" + docsRequired + "), You have: (" + haveDocs + ")"
        )
    } else if (npc.sendYesNo(
        "Exchange (" + docsRequired + ") #t" + itemDocs + "# for (" + scrollReward + ") #t" + itemReturnScroll + "#?"
    )) {
        if (!plr.giveItem(itemReturnScroll, scrollReward)) {
            npc.sendOk("Please make sure you have enough inventory space, then try again.")
        } else {
            var tookD = plr.removeItemsByID(itemDocs, docsRequired)
            if (!tookD) {
                plr.removeItemsByID(itemReturnScroll, scrollReward) // rollback
                npc.sendOk("An error occurred while taking the documents. Please try again.")
            } else {
                npc.sendOk("Exchange complete. Good luck.")
            }
        }
    } else {
        npc.sendOk("Alright. Come back when you are ready.")
    }
}