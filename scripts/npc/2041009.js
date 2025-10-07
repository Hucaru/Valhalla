var couponCut  = 5150006; // EXP Haircut coupon
var couponDye  = 5151006; // Normal Hair Dye coupon

// Main menu
npc.sendSelection(
    "Hey there! I'm the assistant here. If you have #b#t" + couponCut + "##k or #b#t" + couponDye + "##k, I can give you a brand new look!\r\n" +
    "#L0#Haircut (EXP coupon)#l\r\n" +
    "#L1#Dye your hair (normal coupon)#l"
);

var sel = npc.selection();
var z = plr.hair() % 10;

if (sel === 0) {
    // Haircut (random)
    var baseMale = [
        30540, 30640, 30680, 30250, 30190, 30150, 30050,
        30280, 30240, 30300, 30160, 30650
    ];

    var baseFemale = [
        31540, 31640, 31600, 31150, 31280, 31160, 31120,
        31290, 31270, 31030, 31230, 31010, 31680
    ];

    var src = (plr.gender() < 1) ? baseMale : baseFemale;
    var newHair = src[Math.floor(Math.random() * src.length)] + z;

    if (plr.itemCount(couponCut) > 0) {
        plr.removeItemsByID(couponCut, 1);
        plr.setHair(newHair);
        npc.sendOk("Here’s the mirror! What do you think of your new look? I knew it’d turn out great!");
    } else {
        npc.sendOk("Hmm… it seems like you don’t have a #b#t" + couponCut + "##k. Sorry, but I can’t give you a haircut without it!");
    }

} else if (sel === 1) {
    // Hair color (random)
    var base = Math.floor(plr.hair() / 10) * 10;
    var colors = [base, base + 1, base + 2, base + 3, base + 4, base + 5, base + 6, base + 7];
    var newColor = colors[Math.floor(Math.random() * colors.length)];

    if (plr.itemCount(couponDye) > 0) {
        plr.removeItemsByID(couponDye, 1);
        plr.setHair(newColor);
        npc.sendOk("Here’s the mirror! Your new color looks amazing. Don’t forget to stop by again soon!");
    } else {
        npc.sendOk("Hmm… it seems like you don’t have a #b#t" + couponDye + "##k. Sorry, but I can’t dye your hair without it!");
    }
}
