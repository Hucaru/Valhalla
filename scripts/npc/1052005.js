/*
** NPC: Dr. Feeble
** Location: Henesys
** Purpose: Plastic Surgeon (EXP)
*/

const coupon = 5152000; // EXP plastic surgery coupon

// Prompt warning about random result
if (!npc.sendYesNo("If you use the regular coupon, your face will change RANDOMLY, with a chance to obtain a new experimental look I came up with. Do you still want to do it using #b#t" + coupon + "##k?")) {
    npc.sendNext("I see... take your time and see if you really want it. Let me know when you've decided.");
} else if (plr.itemCount(coupon) <= 0) {
    npc.sendNext("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...");
} else {
    // Build random candidate list based on gender
    var base;
    if (plr.gender() == 0) {
        base = [20000, 20005, 20008, 20012, 20016, 20022, 20032];
    } else {
        base = [21000, 21002, 21008, 21014, 21020, 21024, 21029];
    }

    // Keep current color block (hundreds)
    var colorOffset = Math.floor((plr.getFace() / 100) % 10) * 100;

    // Pick random face, then apply color offset
    var face = base[Math.floor(Math.random() * base.length)] + colorOffset;

    // Consume coupon and apply
    plr.removeItemsByID(coupon, 1);
    plr.setFace(face);

    npc.sendNext("Okay, the surgery's done. Here's a mirror—check it out. What a masterpiece, no? Haha! If you ever get tired of this look, please feel free to come visit me again.");
}