npc.sendSelection("I'm Brittany the assistant. If you have #b#t5150052##k or #b#t5151035##k by any chance, then how about letting me change your hairdo? \r\n#L0##bHaircut(REG coupon)#l\r\n#L1#Dye your hair(REG coupon)#l")
var select = npc.selection()

if (select == 0) {
    // Haircut
    var hair
    if (plr.gender() < 1) {
        hair = [30310, 30330, 30060, 30150, 30410, 30210, 30140, 30120, 30200, 30560, 30510, 30610, 30470]
    } else {
        hair = [31150, 31310, 31300, 31160, 31100, 31410, 31030, 31080, 31070, 31610, 31350, 31510, 31740]
    }
    var chosen = hair[Math.floor(Math.random() * hair.length)] + (plr.hair() % 10)

    if (npc.sendYesNo("If you use the REG coupon your hair will change RANDQMLY with a chance to obtain a new experimental style that even you didn't think was possible. Are you going to use #b#t5150052##k and really change your hairstyle?")) {
        if (plr.takeItem(5150052, 1)) {
            plr.setHair(chosen)
            npc.sendBackNext("Hey, here's the mirror. What do you think of your new haircut? I know it wasn't the smoothest of all, but didn't it come out pretty nice? Come back later when you need to change it up again!", true, false)
        } else {
            npc.sendBackNext("Hmmm...are you sure you have our designated coupon? Sorry but no haircut without it.", true, false)
        }
    }
} else if (select == 1) {
    // Dye
    var base = Math.floor(plr.hair() / 10) * 10
    var colors = [base + 0, base + 1, base + 2, base + 3, base + 4, base + 5]
    var chosen = colors[Math.floor(Math.random() * colors.length)]

    if (npc.sendYesNo("If you use a regular coupon your hair will change RANDOMLY. Do you still want to use #b#t5151035##k and change it up?")) {
        if (plr.takeItem(5151035, 1)) {
            plr.setHair(chosen)
            npc.sendBackNext("Hey, here's the mirror. What do you think of your new haircolor? I know it wasn't the smoothest of all, but didn't it come out pretty nice? Come back later when you need to change it up again!", true, false)
        } else {
            npc.sendBackNext("Hmmm...are you sure you have our designated coupon? Sorry but no dye your hair without it.", true, false)
        }
    }
}

// Generate by kimi-k2-instruct