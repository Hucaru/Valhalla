var couponCut  = 5150000; // Normal haircut coupon
var couponDye  = 5151000; // Normal dye coupon

npc.sendSelection(
    "Hi there! I'm the assistant here in Henesys. If you have #b#t" + couponCut + "##k or #b#t" + couponDye + "##k, I can help you change up your look!\r\n" +
    "#L0#Haircut (normal coupon)#l\r\n" +
    "#L1#Dye your hair (normal coupon)#l"
);

var sel = npc.selection();
var z = plr.hair() % 10;

if (sel === 0) {
    var baseMale = [30000, 30020, 30030, 30060, 30120, 30140, 30150, 30200, 30210, 30310, 30330, 30410];
    var baseFemale = [31000, 31030, 31040, 31050, 31080, 31070, 31100, 31150, 31160, 31300, 31310, 31410];
    var src = (plr.gender() < 1) ? baseMale : baseFemale;
    var newHair = src[Math.floor(Math.random() * src.length)] + z;

    if (plr.itemCount(couponCut) > 0) {
        plr.removeItemsByID(couponCut, 1);
        plr.setHair(newHair);
        npc.sendOk("Take a look! I think this new style suits you perfectly!");
    } else {
        npc.sendOk("It seems you don't have a #b#t" + couponCut + "##k. Sorry, I can’t cut your hair without it.");
    }
} else if (sel === 1) {
    var base = Math.floor(plr.hair() / 10) * 10;
    var colors = [base, base + 1, base + 2, base + 3, base + 4, base + 5, base + 6, base + 7];
    var newColor = colors[Math.floor(Math.random() * colors.length)];

    if (plr.itemCount(couponDye) > 0) {
        plr.removeItemsByID(couponDye, 1);
        plr.setHair(newColor);
        npc.sendOk("Your new color looks great! Come back if you ever want another change!");
    } else {
        npc.sendOk("You don't have a #b#t" + couponDye + "##k. Sorry, I can’t dye your hair without it.");
    }
}
