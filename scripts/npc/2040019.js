if (npc.sendYesNo("If you use the regular coupon, your face may transfor into a random new look ... do you still want to do it using #b#t5152056##k?")) {
    var face;
    if (plr.gender() < 1) {
        face = [20001, 20003, 20007, 20013, 20021, 20023, 20025];
    } else {
        face = [21002, 21004, 21006, 21008, 21022, 21027, 21029];
    }
    var newFace = face[Math.floor(Math.random() * face.length)] + Math.floor(plr.face() / 100 % 10) * 100;

    if (plr.takeItem(5152056, 1)) {
        plr.setFace(newFace);
        npc.sendBackNext("Now the procedure's done ... check it out, here's the mirror for you ... what do you think? Even l admit this looks like a masterpiece ... hahah, well, give me a call once you get sick of that new look, alright?", true, true);
    } else {
        npc.sendBackNext("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...", true, true);
    }
} else {
    npc.sendBackNext("Come back anytime if you change your mind.", true, true);
}

// Generate by kimi-k2-instruct