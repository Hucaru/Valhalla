var couponCut  = 5150001; // Haircut coupon
var couponDye  = 5151001; // Hair dye coupon

// Main menu
npc.sendSelection(
    "I'm the head of this hair salon Natalie. If you have #b#t" + couponCut + "##k or #b#t" + couponDye + "##k, allow me to take care of your hairdo. Please choose the one you want.\r\n"
    + "#L0#Haircut (VIP coupon)#l\r\n"
    + "#L1#Dye your hair (VIP coupon)#l"
);
var sel = npc.selection();

var z = plr.hair() % 10;

if (sel === 0) {
    // Haircut branch — build from base IDs + current color digit
    var baseMale   = [30030, 30020, 30000, 30060, 30150, 30210, 30140, 30120, 30200, 30170];
    var baseFemale = [31050, 31040, 31000, 31150, 31160, 31100, 31030, 31080, 31030, 31070];

    var hair = [];
    var src = (plr.gender() < 1) ? baseMale : baseFemale;
    for (var i = 0; i < src.length; i++) {
        hair.push(src[i] + z);
    }

    npc.sendAvatar.apply(npc,
        ["I can totally change up your hairstyle and make it look so good. Why don't you change it up a bit? With #b#t" + couponCut + "##k I'll change it for you. Choose the one to your liking~"]
            .concat(hair)
    );
    var choice = npc.selection();

    if (choice < 0 || choice >= hair.length) {
        npc.sendOk("Changed your mind? That's fine. Come back any time.");
    } else if (plr.itemCount(couponCut) > 0) {
        plr.removeItemsByID(couponCut, 1);
        plr.setHair(hair[choice]);
        npc.sendBackNext(
            "Check it out!! What do you think? Even I think this one is a work of art! AHAHAHA. Please let me know when you want another haircut, because I'll make you look good each time!",
            false, true
        );
    } else {
        npc.sendBackNext(
            "Hmmm... it looks like you don't have our designated coupon... I'm afraid I can't give you a haircut without it. I'm sorry...",
            false, true
        );
    }

} else if (sel === 1) {
    // Hair dye branch — keep style, vary color 0..7
    var base = Math.floor(plr.hair() / 10) * 10;
    var colors = [base + 0, base + 1, base + 2, base + 3, base + 4, base + 5, base + 6, base + 7];

    npc.sendAvatar.apply(npc,
        ["I can totally change your haircolor and make it look so good. Why don't you change it up a bit? With #b#t" + couponDye + "##k I'll change it for you. Choose the one to your liking."]
            .concat(colors)
    );
    var choice = npc.selection();

    if (choice < 0 || choice >= colors.length) {
        npc.sendOk("Changed your mind? That's fine. Come back any time.");
    } else if (plr.itemCount(couponDye) > 0) {
        plr.removeItemsByID(couponDye, 1);
        plr.setHair(colors[choice]);
        npc.sendBackNext(
            "Check it out!! What do you think? Even I think this one is a work of art! AHAHAHA. Please let me know when you want to dye your hair again, because I'll make you look good each time!",
            false, true
        );
    } else {
        npc.sendBackNext(
            "Hmmm... it looks like you don't have our designated coupon... I'm afraid I can't dye your hair without it. I'm sorry...",
            false, true
        );
    }
}