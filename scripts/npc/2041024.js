var minLevel = 60

var qStart = 7100          // Protect Ludibrium
var qPrev = 7106           // The Lost Medal (prereq for 7107)
var qCrackedPiece = 7107   // The Lost Piece of Crack (repeatable handoff at Flo)

var crackedPieceItem = 4031179
var finishMap7100 = 220050300
var finishMap7107 = 220050300

var startStatus = plr.getQuestStatus(qStart)          // 2 = completed, 1 = in progress, 0 = not started
var prevStatus = plr.getQuestStatus(qPrev)            // 2/1/0
var pieceStatus = plr.getQuestStatus(qCrackedPiece)   // 2/1/0

// Offer 7100 if eligible and not started
if (plr.level() >= minLevel && startStatus == 0) {
    if (npc.sendYesNo("Brave Adventurer, you've reached level 60!\r\nWould you like to start the quest #bProtect Ludibrium#k?")) {
        plr.startQuest(qStart)
        npc.sendOk("Great! Please go see Mr. Bouffon. I'll send you there now.")
        plr.warp(finishMap7100)
    } else {
        npc.sendOk("Very well. Speak to me again if you change your mind.")
    }
} else if (plr.level() < minLevel || prevStatus != 2) {
    // Default gate text if under-leveled or prerequisite (7106) not complete yet
    npc.sendOk("For those capable of great feats and bearers of an unwavering resolve, the #bfinal destination#k lies ahead past the gate. The Machine Room accepts only #rone party at a time#k, so make sure your party is ready when crossing the gate.")
} else if (pieceStatus == 0) {
    // 7106 complete and 7107 not started — offer to start and warp to Flo
    if (npc.sendYesNo("Would you like to start this quest to receive the #t" + crackedPieceItem + "# and fight Papulatus?")) {
        plr.startQuest(qCrackedPiece)
        npc.sendOk("Head to Flo to complete it. I'll send you there now.")
        plr.warp(finishMap7107)
    } else {
        npc.sendOk("Alright. Come back when you are ready.")
    }
} else if (pieceStatus == 1) {
    // 7107 in progress — offer to warp to Flo to finish
    if (npc.sendYesNo("You're already on this request.\r\nWould you like me to send you to Flo to finish it?")) {
        npc.sendOk("Good luck.")
        plr.warp(finishMap7107)
    } else {
        npc.sendOk("Very well. Proceed when you are ready.")
    }
} else {
    // 7107 completed — offer to restart (reset then start) and warp to Flo
    if (npc.sendYesNo("Would you like to restart this quest to receive the #t" + crackedPieceItem + "# again and fight Papulatus?")) {
        plr.setQuestData(qCrackedPiece, "")
        plr.startQuest(qCrackedPiece)
        npc.sendOk("The request has been started again. I'll send you to Flo now.")
        plr.warp(finishMap7107)
    } else {
        npc.sendOk("Understood. Return when you wish to try again.")
    }
}