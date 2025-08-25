npc.sendSelection("Welcome, welcome, welcome to the Ludibrium Hair-Salon! Do you, by any chance, have #b#t5150053##k or #b#t5151036##k? If so, how about letting me take care of your hair? Please choose what you want to do with it. \r\n#L0##bChange hair style (VIP coupon)#l\r\n#L1#Dye your hair (VIP coupon)#l")
var select = npc.selection()

if (select == 0) {
    // Change hair style
    var hair
    if (plr.gender() < 1) {
        hair = [30250, 30190, 30660, 30870, 30840, 30990, 30160, 30640]
    } else {
        hair = [31810, 31550, 31830, 31840, 31680, 31290, 31270, 31870]
    }
    for (var i = 0; i < hair.length; i++) {
        hair[i] = hair[i] + (plr.hair() % 10)
    }
    npc.sendStyles("I can completely change the look of your hair. Aren't you ready for a change? With #b#t5150053##k, I'll take care of the rest for you. Choose the style of your liking!", hair)
    var styleSel = npc.selection()
    if (plr.takeItem(5150053, 1)) {
        plr.setHair(hair[styleSel])
        npc.sendBackNext("Enjoy your new and improved hairstyle!", true, true)
    } else {
        npc.sendBackNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't give you a haircut without it. I'm sorry...", true, true)
    }
} else if (select == 1) {
    // Dye hair
    var base = Math.floor(plr.hair() / 10) * 10
    var hair = [base + 0, base + 2, base + 3, base + 4, base + 5]
    npc.sendStyles("I can completely change the color of your hair. Aren't you ready for a change? With #b#t5151036##k, I'll take care of the rest. Choose the color of your liking!", hair)
    var colorSel = npc.selection()
    if (plr.takeItem(5151036, 1)) {
        plr.setHair(hair[colorSel])
        npc.sendBackNext("Enjoy your new and improved hair colour!", true, true)
    } else {
        npc.sendBackNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't dye your hair without it. I'm sorry...", true, true)
    }
}

// Generate by kimi-k2-instruct