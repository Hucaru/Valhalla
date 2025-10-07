var couponCut  = 5150002;
var couponDye  = 5151002;

npc.sendSelection(
    "Yo! I'm the assistant at the Kerning salon. Got #b#t" + couponCut + "##k or #b#t" + couponDye + "##k? Let's give you a new look!\r\n" +
    "#L0#Haircut (normal coupon)#l\r\n" +
    "#L1#Dye your hair (normal coupon)#l"
);

var sel = npc.selection();
var z = plr.hair() % 10;

if (sel === 0) {
    var baseMale = [30000, 30020, 30030, 30040, 30050, 30110, 30130, 30160, 30180, 30190, 30350, 30610, 30440, 30400];
    var baseFemale = [31000, 31010, 31020, 31040, 31050, 31060, 31090, 31120, 31130, 31140, 31330, 31700, 31620, 31610];
    var src = (plr.gender() < 1) ? baseMale : baseFemale;
    var newHair = src[Math.floor(Math.random() * src.length)] + z;

    if (plr.itemCount(couponCut) > 0) {
        plr.removeItemsByID(couponCut, 1);
        plr.setHair(newHair);
        npc.sendOk("All done! What do you think? A fresh new cut for a fresh new start!");
    } else {
        npc.sendOk("Looks like you're missing a #b#t" + couponCut + "##k. Sorry, I can’t do it without that!");
    }
} else if (sel === 1) {
    var base = Math.floor(plr.hair() / 10) * 10;
    var colors = [base, base + 1, base + 2, base + 3, base + 4, base + 5, base + 6, base + 7];
    var newColor = colors[Math.floor(Math.random() * colors.length)];

    if (plr.itemCount(couponDye) > 0) {
        plr.removeItemsByID(couponDye, 1);
        plr.setHair(newColor);
        npc.sendOk("Your new color is poppin'! Come back anytime!");
    } else {
        npc.sendOk("No #b#t" + couponDye + "##k, no color. Sorry, friend!");
    }
}
