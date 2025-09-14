// beauty coupon dialog
if (npc.sendYesNo("If you use the regular coupon, your face may transform into a random new look ... do you still want to do it using #b#t5152056##k?")) {
    if (plr.itemCount(5152056) > 0) {
        var face;
        if (plr.job() < 1 || (plr.job() >= 1000 && plr.job() < 2000)) {
            var maleChoices = [20001, 20003, 20007, 20013, 20021, 20023, 20025];
            face = maleChoices[Math.floor(Math.random() * maleChoices.length)] + parseInt(plr.job() / 100 % 10) * 100;
        } else {
            var femaleChoices = [21002, 21004, 21006, 21008, 21022, 21027, 21029];
            face = femaleChoices[Math.floor(Math.random() * femaleChoices.length)] + parseInt(plr.job() / 100 % 10) * 100;
        }
        
        plr.removeItemsByID(5152056, 1);
        plr.setFace(face);
        plr.updateSingleStat(Packages.client.MapleStat.FACE, face);
        npc.sendOk("Now the procedure's done ... check it out, here's the mirror for you ... what do you think? Even I admit this looks like a masterpiece ... hahah, well, give me a call once you get sick of that new look, alright?");
    } else {
        npc.sendOk("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...");
    }
}