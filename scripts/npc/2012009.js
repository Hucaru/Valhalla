// Face surgery assistant 莉茲
if (npc.sendYesNo("If you use the regular coupon, your face may be changed to something random. Do you still want to proceed using #b#t5152056##k?")) {
    if (plr.itemCount(5152056) > 0) {
        plr.takeItem(5152056, -1, 1, 1); // remove coupon
        var face;
        if (plr.job() < 1000) { // male
            face = [20003, 20011, 20021, 20022, 20023, 20027, 20031];
        } else { // female
            face = [21004, 21007, 21010, 21012, 21020, 21021, 21030];
        }
        var chosenFace = face[Math.floor(Math.random() * face.length)] + Math.floor(plr.getFace() / 100 % 10) * 100;
        plr.setFace(chosenFace); // sets face directly
        npc.sendNext("The procedure's done...Check it out in the mirror. What do you think? Simply beautiful... Haha! Well, give me a call if you get sick of it, alright?");
    } else {
        npc.sendNext("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...");
    }
}