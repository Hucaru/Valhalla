// Minu (Ludibrium Hair-Salon)
var choice = npc.sendBackNext(
    "Welcome, welcome, welcome to the Ludibrium Hair-Salon! Do you, by any chance, have #b#t5150053##k or #b#t5151036##k? If so, how about letting me take care of your hair? Please choose what you want to do with it. \r\n#L0##bChange hair style (VIP coupon)#l\r\n#L1#Dye your hair (VIP coupon)#l",
    false, true);

if (choice === 0) {      // Change style
    var base = (plr.job() == 0 ? 30000 : 31000);           // Male vs Female base
    var hair = [
        base + 250, base + 190, base + 660, base + 870, base + 840,
        base + 990, base + 160, base + 640
    ];
    /* add current hair-color digit (last digit) */
    var color = plr.getHair() % 10;
    for (var i = 0; i < hair.length; i++)
        hair[i] += color;

    var pick = npc.sendAvatar("I can completely change the look of your hair. Aren't you ready for a change? With #b#t5150053##k, I'll take care of the rest for you. Choose the style of your liking!", hair);

    if (plr.itemCount(5150053)) {
        plr.inventoryExchange(5150053, 1, hair[pick], 1);
        npc.sendNext("Enjoy your new and improved hairstyle!");
    } else {
        npc.sendNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't give you a haircut without it. I'm sorry...");
    }

} else if (choice === 1) { // Dye hair
    var base = Math.floor(plr.getHair() / 10) * 10;
    var hair = [base + 0, base + 2, base + 3, base + 4, base + 5];

    var pick = npc.sendAvatar(
        "I can completely change the color of your hair. Aren't you ready for a change? With #b#t5151036##k, I'll take care of the rest. Choose the color of your liking!",
        hair);

    if (plr.itemCount(5151036)) {
        plr.inventoryExchange(5151036, 1, hair[pick], 1);
        npc.sendNext("Enjoy your new and improved hair colour!");
    } else {
        npc.sendNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't dye your hair without it. I'm sorry...");
    }
}