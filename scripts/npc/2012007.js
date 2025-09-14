// Rinz the Assistant — stateless version
var regCouponHaircut = 5150052;
var regCouponDye = 5151035;

// Main intro
var sel = npc.menu(
    "I'm Rinz the Assistant, the assistant. Do you have #bOrbis Hair Salon (Reg. Coupon)#k? If so, what do you think about letting me take care of your hair do? What do you want to do with your hair?",
    "Haircut(reg coupon)",
    "Dye your hair(reg coupon)"
);

// Map the gender to available base styles
var baseStyles = (plr.gender() === 0)
    ? [30030, 30020, 30000, 30270, 30230, 30260, 30280, 30240, 30290, 30340, 30370, 30630, 30530, 30760]
    : [31040, 31000, 31250, 31220, 31260, 31240, 31110, 31270, 31030, 31230, 31530, 31710, 31320, 31650, 31630];

switch (sel) {
    case 0: // Haircut
        var newHair = baseStyles[Math.floor(Math.random() * baseStyles.length)] + (plr.getHair() % 10);
        if (npc.sendYesNo("If you use the regular coupon your hairstyle will randomly change. Do you want to use #b#t" + regCouponHaircut + "##k and change your hair?")) {
            if (plr.itemCount(regCouponHaircut) > 0) {
                plr.takeItem(regCouponHaircut, 0, 1, 4);
                plr.setHair(newHair);
                npc.sendNext("Hey, here's the mirror. What do you think of your new haircut? I know it wasn't the smoothest of all, but didn't it come out pretty nice? If you ever feel like changing it up again later, please drop by.");
            } else {
                npc.sendNext("Hmmm...are you sure you have our designated coupon? Sorry but no haircut without it.");
            }
        }
        break;

    case 1: // Dye
        var base = Math.floor(plr.getHair() / 10) * 10;
        var tones = [base, base + 1, base + 2, base + 3, base + 4, base + 5];
        var newHair = tones[Math.floor(Math.random() * tones.length)];
        if (npc.sendYesNo("If you use the regular coupon your haircolor will randomly change. Do you still want to use #b#t" + regCouponDye + "##k and change it?")) {
            if (plr.itemCount(regCouponDye) > 0) {
                plr.takeItem(regCouponDye, 0, 1, 4);
                plr.setHair(newHair);
                npc.sendNext("Hey, here's the mirror. What do you think of your new haircolor? I know it wasn't the smoothest of all, but didn't it come out pretty nice? If you ever feel like changing it up again later, please drop by.");
            } else {
                npc.sendNext("Hmmm...are you sure you have our designated coupon? Sorry but no dye your hair without it.");
            }
        }
        break;
}