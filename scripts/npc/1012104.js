// Brittany the assistant hair NPC – REG coupons (random result)
const couponHaircut = 5150000
const couponDye = 5151000

// Use base hairstyle IDs that are valid on this version, then add the current color digit
const maleBase   = [30030, 30020, 30000, 30060, 30150, 30210, 30140, 30120, 30200, 30170]
const femaleBase = [31050, 31040, 31000, 31150, 31160, 31100, 31030, 31080, 31030, 31070]

// Intro
npc.sendBackNext(
    "I'm Brittany the assistant. If you have #b#t" + couponHaircut + "##k or #b#t" + couponDye + "##k by any chance, then how about letting me change your hairdo?",
    false, true
)

// Menu
npc.sendSelection(
    "What would you like to do today?\r\n" +
    "#L0##bHaircut (REG coupon)#k#l\r\n" +
    "#L1##bDye your hair (REG coupon)#k#l"
)
var choice = npc.selection()

if (choice === 0) {
    // Haircut (random style within gender, keep color)
    var z = plr.hair() % 10
    var basePool = (plr.gender() < 1) ? maleBase : femaleBase
    var newStyle = basePool[Math.floor(Math.random() * basePool.length)] + z

    if (!npc.sendYesNo("If you use the REG coupon, your hair will change RANDOMLY. Use #b#t" + couponHaircut + "##k and change your hairstyle?")) {
        npc.sendOk("See you another time!")
    } else if (plr.itemCount(couponHaircut) >= 1) {
        plr.removeItemsByID(couponHaircut, 1)
        plr.setHair(newStyle)
        npc.sendOk("Hey, here's the mirror. What do you think of your new haircut? Come back later when you want to change it up again!")
    } else {
        npc.sendOk("Hmmm... are you sure you have our designated coupon? Sorry, no haircut without it.")
    }

} else if (choice === 1) {
    // Dye (random color, keep base style)
    var base = Math.floor(plr.hair() / 10) * 10
    // Offer full 0..7 color range for compatibility
    var colors = [base + 0, base + 1, base + 2, base + 3, base + 4, base + 5, base + 6, base + 7]
    var newStyle = colors[Math.floor(Math.random() * colors.length)]

    if (!npc.sendYesNo("If you use the REG coupon, your hair color will change RANDOMLY. Use #b#t" + couponDye + "##k and change it up?")) {
        npc.sendOk("See you another time!")
    } else if (plr.itemCount(couponDye) >= 1) {
        plr.removeItemsByID(couponDye, 1)
        plr.setHair(newStyle)
        npc.sendOk("Hey, here's the mirror. What do you think of your new hair color? Come back later when you want to change it up again!")
    } else {
        npc.sendOk("Hmmm... are you sure you have our designated coupon? Sorry, we can’t dye your hair without it.")
    }

} else {
    npc.sendOk("See you another time!")
}