if (npc.sendYesNo("If you use the regular coupon, your face may be changed to something random. Do you still want to proceed using #b#t5152056##k?")) {
    var face;
    if (plr.gender() < 1) {
        face = [20003, 20011, 20021, 20022, 20023, 20027, 20031];
    } else {
        face = [21004, 21007, 21010, 21012, 21020, 21021, 21030];
    }
    var newFace = face[Math.floor(Math.random() * face.length)] + Math.floor(plr.face() / 100 % 10) * 100;

    if (plr.takeItem(5152056, 1)) {
        plr.setFace(newFace);
        npc.sendOk("The procedure's done...Check it out in the mirror. What do you think? Simply beautiful... Haha! Well, give me a call if you get sick of it, alright?");
    } else {
        npc.sendOk("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...");
    }
}

// Generate by kimi-k2-instruct