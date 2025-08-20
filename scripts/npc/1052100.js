// Don Giovanni – beauty salon
var menu = "Hello! I'm Don Giovanni, head of the beauty salon! If you have either #b#t5150053##k or #b#t5151036##k, why don't you let me take care of the rest? Decide what you want to do with your hair... \r\n#L0##bChange hair style (VIP coupon)#l\r\n#L1#Dye your hair (VIP coupon)#l"
npc.sendSelection(menu)
var sel = npc.selection()

if (sel == 0) {
    // Change hair style
    var hair
    if (plr.gender() < 1) {
        hair = [30130, 33040, 30850, 30780, 30040, 30920]
    } else {
        hair = [34090, 31090, 31880, 31140, 31330, 31760]
    }
    for (var i = 0; i < hair.length; i++) {
        hair[i] = hair[i] + (plr.hair() % 10)
    }
    npc.sendStyles("I can change your hairstyle to something totally new. Aren't you sick of your hairdo? I'll give you a haircut with #b#t5150053##k. Choose the hairstyle of your liking.", hair)
    var pick = npc.selection()
    if (plr.takeItem(5150053, 1)) {
        plr.setHair(hair[pick])
        npc.sendBackNext("Ok, check out your new haircut. What do you think? Even I admit this one is a masterpiece! AHAHAHA. Let me know when you want another haircut. I'll take care of the rest!", true, true)
    } else {
        npc.sendBackNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't give you a haircut without it. I'm sorry...", true, true)
    }
} else if (sel == 1) {
    // Dye hair
    var base = Math.floor(plr.hair() / 10) * 10
    var hair = [base + 0, base + 2, base + 3, base + 5]
    npc.sendStyles("I can change the color of your hair to something totally new. Aren't you sick of your hair-color? I'll dye your hair if you have #bVIP hair color coupon#k. Choose the hair-color of your liking!", hair)
    var pick = npc.selection()
    if (plr.takeItem(5151036, 1)) {
        plr.setHair(hair[pick])
        npc.sendBackNext("Ok, check out your new hair color. What do you think? Even I admit this one is a masterpiece! AHAHAHA. Let me know when you want another haircut. I'll take care of the rest!", true, true)
    } else {
        npc.sendBackNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't dye your hair without it. I'm sorry...", true, true)
    }
}

// Generate by kimi-k2-instruct