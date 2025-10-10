var couponCut  = 5150004;
var couponDye  = 5151004;

npc.sendSelection(
    "Welcome to Orbis Hair! If you have #b#t" + couponCut + "##k or #b#t" + couponDye + "##k, I can take care of your hair!\r\n" +
    "#L0#Haircut (normal coupon)#l\r\n" +
    "#L1#Dye your hair (normal coupon)#l"
);

var sel = npc.selection();
var z = plr.hair() % 10;

if (sel === 0) {
    var baseMale = [30000, 30020, 30030, 30230, 30240, 30260, 30270, 30280, 30290, 30340, 30610, 30440, 30400];
    var baseFemale = [31000, 31030, 31040, 31110, 31220, 31230, 31240, 31250, 31260, 31270, 31320, 31700, 31620, 31610];
    var src = (plr.gender() < 1) ? baseMale : baseFemale;
    var newHair = src[Math.floor(Math.random() * src.length)] + z;

    if (plr.itemCount(couponCut) > 0) {
        plr.removeItemsByID(couponCut, 1);
        plr.setHair(newHair);
        npc.sendOk("There we go! What do you think? A new style for a new you!");
    } else {
        npc.sendOk("It looks like you don't have a #b#t" + couponCut + "##k. Sorry, can't help without it!");
    }
} else if (sel === 1) {
    var base = Math.floor(plr.hair() / 10) * 10;
    var colors = [base, base + 1, base + 2, base + 3, base + 4, base + 5, base + 6, base + 7];
    var newColor = colors[Math.floor(Math.random() * colors.length)];

    if (plr.itemCount(couponDye) > 0) {
        plr.removeItemsByID(couponDye, 1);
        plr.setHair(newColor);
        npc.sendOk("Beautiful! Your new color shines brighter than the clouds themselves!");
    } else {
        npc.sendOk("No #b#t" + couponDye + "##k, no dye job. Sorry!");
    }
}
