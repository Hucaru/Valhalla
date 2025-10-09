var couponCut  = 5150007; // Haircut coupon
var couponDye  = 5151007; // Hair dye coupon

// Main menu
npc.sendSelection(
    "I'm the head of this hair salon. If you have #b#t" + couponCut + "##k or #b#t" + couponDye + "##k, allow me to take care of your hairdo. Please choose the one you want.\r\n"
    + "#L0#Haircut (VIP coupon)#l\r\n"
    + "#L1#Dye your hair (VIP coupon)#l"
);
var sel = npc.selection();

var z = plr.hair() % 10;

if (sel === 0) {
    var baseMale = [
        30030, 30020, 30000, 30660, 30250, 30190, 30150, 30050, 30280, 30240, 30300, 30160
    ];

    var baseFemale = [
        31040, 31000, 31550, 31150, 31280, 31160, 31120, 31290, 31270, 31030, 31230, 31010
    ];

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
        npc.sendOk(
            "Check it out!! What do you think? Even I think this one is a work of art! AHAHAHA. Please let me know when you want another haircut, because I'll make you look good each time!");
    } else {
        npc.sendOk(
            "Hmmm... it looks like you don't have our designated coupon... I'm afraid I can't give you a haircut without it. I'm sorry..."
        );
    }

} else if (sel === 1) {
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
        npc.sendOk(
            "Check it out!! What do you think? Even I think this one is a work of art! AHAHAHA. Please let me know when you want to dye your hair again, because I'll make you look good each time!"
        );
    } else {
        npc.sendOk(
            "Hmmm... it looks like you don't have our designated coupon... I'm afraid I can't dye your hair without it. I'm sorry..."
        );
    }
}