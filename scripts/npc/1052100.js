// Don Giovanni – stateless beauty-salon script

// Greeting
npc.sendSelection(
    "Hello! I'm Don Giovanni, head of the beauty salon! If you have either #b#t5150053##k or #b#t5151036##k, why don't you let me take care of the rest? Decide what you want to do with your hair...\r\n#L0##bChange hair style (VIP coupon)#l\r\n#L1#Dye your hair (VIP coupon)#l"
);
var sel = npc.selection();

if (sel === 0) {
    // Change hairstyle
    var gender = plr.job() < 2000 ? 0 : 1;
    var base = gender === 0
        ? [30130, 33040, 30850, 30780, 30040, 30920]
        : [34090, 31090, 31880, 31140, 31330, 31760];

    for (var i = 0; i < base.length; i++) {
        base[i] = base[i] + plr.getHair() % 10;
    }

    var choice = npc.sendAvatar(
        "I can change your hairstyle to something totally new. Aren't you sick of your hairdo? I'll give you a haircut with #b#t5150053##k. Choose the hairstyle of your liking.",
        base
    );

    if (plr.itemCount(5150053)) {
        plr.giveItem(5150053, -1);
        plr.setHair(base[choice]);
        npc.sendOk("Ok, check out your new haircut. What do you think? Even I admit this one is a masterpiece! AHAHAHA. Let me know when you want another haircut. I'll take care of the rest!");
    } else {
        npc.sendOk("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't give you a haircut without it. I'm sorry...");
    }

} else if (sel === 1) {
    // Dye hair
    var hairBase = Math.floor(plr.getHair() / 10) * 10;
    var colors = [hairBase + 0, hairBase + 2, hairBase + 3, hairBase + 5];

    var choice = npc.sendAvatar(
        "I can change the color of your hair to something totally new. Aren't you sick of your hair-color? I'll dye your hair if you have #bVIP hair color coupon#k. Choose the hair-color of your liking!",
        colors
    );

    if (plr.itemCount(5151036)) {
        plr.giveItem(5151036, -1);
        plr.setHair(colors[choice]);
        npc.sendOk("Ok, check out your new hair color. What do you think? Even I admit this one is a masterpiece! AHAHAHA. Let me know when you want another haircut. I'll take care of the rest!");
    } else {
        npc.sendOk("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't dye your hair without it. I'm sorry...");
    }
}