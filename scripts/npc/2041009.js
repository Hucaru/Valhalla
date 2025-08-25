npc.sendSelection("Hi, I'm the assistant here. Don't worry, I'm plenty good enough for this. If you have #b#t5150052##k or #b#t5151035##k by any chance, then allow me to take care of the rest, alright? \r\n#L0##bChange hair-style (Regular coupon)#l\r\n#L1#Dye your hair (Regular coupon)#l")
var select = npc.selection()

if (select == 0) {
    // Change hair-style
    var hair
    if (plr.gender() < 1) {
        hair = [30250, 30190, 30150, 30050, 30280, 30240, 30300, 30160, 30650, 30540, 30640, 30680]
    } else {
        hair = [31150, 31280, 31160, 31120, 31290, 31270, 31030, 31230, 31010, 31640, 31540, 31680, 31600]
    }
    var chosen = hair[Math.floor(Math.random() * hair.length)] + (plr.hair() % 10)

    if (npc.sendYesNo("If you use the regular coupon, your hair-style will be changed into a random new look. Are you sure you want to use #b#t5150052##k and change it?")) {
        if (plr.takeItem(5150052, 1)) {
            plr.setHair(chosen)
            npc.sendBackNext("Hey, here's the mirror. What do you think of your new haircut? I know it wasn't the smoothest of all, but didn't it come out pretty nice? If you ever feel like changing it up again later, please drop by.", true, true)
        } else {
            npc.sendBackNext("Hmmm...are you sure you have our designated coupon? Sorry but no haircut without it.", true, true)
        }
    }
} else if (select == 1) {
    // Dye hair
    var base = Math.floor(plr.hair() / 10) * 10
    var colors = [base + 0, base + 1, base + 2, base + 3, base + 4, base + 5]
    var chosen = colors[Math.floor(Math.random() * colors.length)]

    if (npc.sendYesNo("If you use the regular coupon, your hair-color will be changed into a random new look. Are you sure you want to use #b#t5151035##k and change it?")) {
        if (plr.takeItem(5151035, 1)) {
            plr.setHair(chosen)
            npc.sendBackNext("Hey, here's the mirror. What do you think of your new haircolor? I know it wasn't the smoothest of all, but didn't it come out pretty nice? If you ever feel like changing it up again later, please drop by.", true, true)
        } else {
            npc.sendBackNext("Hmmm...are you sure you have our designated coupon? Sorry but no dye your hair without it.", true, true)
        }
    }
}

// Generate by kimi-k2-instruct