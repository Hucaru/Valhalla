npc.sendSelection("Hello I'm Mino the Owner. If you have either #b#t5150053##k or #b#t5151036##k, then please let me take care of your hair. Choose what you want to do with it. \r\n#L0##bHaircut(VIP coupon)#l\r\n#L1#Dye your hair(VIP coupon)#l");
var sel = npc.selection();

if (sel === 0) { // Haircut
    var hair;
    if (plr.job() < 1000) { // 0 = male, 1 = female
        hair = [33240, 30230, 30490, 30260, 30280, 33050, 30340];
    } else {
        hair = [34060, 31220, 31110, 31790, 31230, 31630, 34260];
    }
    for (var i = 0; i < hair.length; i++) {
        hair[i] = hair[i] + (plr.getHair() % 10);
    }

    var choice = npc.sendAvatar("I can completely change the look of your hair. Are you ready for a change by any chance? With #b#t5150053##k I can change it up for you. please choose the one you want!", hair);
    if (plr.itemCount(5150053) > 0) {
        plr.removeItemsByID(5150053, 1);
        plr.setHair(hair[choice]);
        npc.sendOk("Enjoy your new and improved hairstyle!");
    } else {
        npc.sendOk("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't give you a haircut without it. I'm sorry...");
    }
} else { // Dye
    var base = Math.floor(plr.getHair() / 10) * 10;
    var dyes = [base + 0, base + 1, base + 3, base + 4, base + 5];

    var dyeChoice = npc.sendAvatar("I can completely change your haircolor. Are you ready for a change by any chance? With #b#t5151036##k I can change it up for you. Please choose the one you want!", dyes);
    if (plr.itemCount(5151036) > 0) {
        plr.removeItemsByID(5151036, 1);
        plr.setHair(dyes[dyeChoice]);
        npc.sendOk("Enjoy your new and improved hair colour!");
    } else {
        npc.sendOk("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't dye your hair without it. I'm sorry...");
    }
}