// I'm Rinz the Assistant
var menu = "I'm Rinz the Assistant, the assistant. Do you have #bOrbis Hair Salon (Reg. Coupon)#k? If so, what do you think about letting me take care of your hair do? What do you want to do with your hair? \r\n#L0##bHaircut(reg coupon)#l\r\n#L1#Dye your hair(reg coupon)#l"
npc.sendSelection(menu)
var select = npc.selection()

if (select == 0) {
    // Haircut
    var hair
    if (plr.gender() < 1) {
        var maleHair = [30030, 30020, 30000, 30270, 30230, 30260, 30280, 30240, 30290, 30340, 30370, 30630, 30530, 30760]
        hair = maleHair[Math.floor(Math.random() * maleHair.length)] + (plr.hair() % 10)
    } else {
        var femaleHair = [31040, 31000, 31250, 31220, 31260, 31240, 31110, 31270, 31030, 31230, 31530, 31710, 31320, 31650, 31630]
        hair = femaleHair[Math.floor(Math.random() * femaleHair.length)] + (plr.hair() % 10)
    }

    if (npc.sendYesNo("If you use the regular coupon your hairstyle with randomly change. Do you want to use #b#t5150052##k and change your hair?")) {
        if (plr.takeItem(5150052, 1)) {
            plr.setHair(hair)
            npc.sendBackNext("Hey, here's the mirror. What do you think of your new haircut? I know it wasn't the smoothest of all, but didn't it come out pretty nice? If you ever feel like changing it up again later, please drop by.", true, true)
        } else {
            npc.sendBackNext("Hmmm...are you sure you have our designated coupon? Sorry but no haircut without it.", true, true)
        }
    }
} else if (select == 1) {
    // Dye hair
    var base = Math.floor(plr.hair() / 10) * 10
    var colors = [base + 0, base + 1, base + 2, base + 3, base + 4, base + 5]
    var hair = colors[Math.floor(Math.random() * colors.length)]

    if (npc.sendYesNo("If you use the regular coupon your haircolor will randomly change. Do you still want to use #b#t5151035##k and change it?")) {
        if (plr.takeItem(5151035, 1)) {
            plr.setHair(hair)
            npc.sendBackNext("Hey, here's the mirror. What do you think of your new haircolor? I know it wasn't the smoothest of all, but didn't it come out pretty nice? If you ever feel like changing it up again later, please drop by.", true, true)
        } else {
            npc.sendBackNext("Hmmm...are you sure you have our designated coupon? Sorry but no dye your hair without it.", true, true)
        }
    }
}

// Generate by kimi-k2-instruct