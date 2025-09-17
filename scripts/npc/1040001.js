if (plr.getQuestStatus(2048) != 1) {
    npc.sendOk("This is a dangerous place. Do be careful...");
} else {
    npc.sendNext("Hmmm... so you want to know how to get #bPiece of Ice#k, #bAncient Scroll#k, and #bFlaming Feather#k? What do you plan on doing with those precious materials? I've heard about those, since I studied the island a bit before doing my work now as a guard...");
    npc.sendBackNext("#bPiece of Ice#k, #bAncient Scroll#k, and #bFlaming Feather#k... those items should be available around the island. If you're looking for an Ancient Scroll, I was told that the ancient magicians used them to create the Golems. Perhaps they'd have it...", true, true)
    npc.sendBackNext("A #bPiece of Ice#k that never melts... The fairies had them... I hear Ice Drakes have them too...", true, true)
    npc.sendBackNext("#bFlaming Feather#k ... I've heard of that, a feather-resembling flame ... it has something to do with a flame-blowing dragon or something ... anyway it's unbelievably vicious, so it'll be difficult for you to take Flaming Feather away from it. Good luck.", true, true)
    npc.sendBackNext("This is a dangerous place. Do be careful...", true, false)
}