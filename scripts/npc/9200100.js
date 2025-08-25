npc.sendSelection("Why hello there! I'm Dr. Lenu, in charge of the cosmetic lenses here at the Henesys Plastic Surgery Shop! With #b#t5152010##k or #b#t5152013##k, you can have the kind of look you've always wanted! All you have to do is find the cosmetic lens that most fits you, then let us take care of the rest. Now, what would you like to use? \r\n#L0##bCosmetic Lenses at Henesys (Reg coupon)#l\r\n#L1#Cosmetic Lenses at Henesys (VIP coupon)#l")
var select = npc.selection()

if (select == 0) {
    // Regular coupon path
    var color = [100, 200, 300, 400, 500, 600, 700]
    var teye = plr.face() % 100
    teye += plr.gender() < 1 ? 20000 : 21000
    color = color[Math.floor(Math.random() * color.length)] + teye

    if (npc.sendYesNo("If you use the regular coupon, you'll be awarded a random pair of cosmetic lenses. Are you going to use #b#t5152010##k and really make the change to your eyes?")) {
        if (plr.takeItem(5152010, 1)) {
            plr.setFace(color)
            npc.sendBackNext("Tada~! Check it out!! What do you think? I really think your eyes look sooo fantastic now~~! Please come again ~", true, true)
        } else {
            npc.sendBackNext("I'm sorry, but I don't think you have our cosmetic lens coupon with you right now. Without the coupon, I'm afraid I can't do it for you..", true, true)
        }
    }
} else if (select == 1) {
    // VIP coupon path
    var color = [100, 200, 300, 400, 500, 600, 700]
    var teye = plr.face() % 100
    teye += plr.gender() < 1 ? 20000 : 21000

    for (var i = 0; i < color.length; i++) {
        color[i] = teye + color[i]
    }

    npc.sendStyles("With our specialized machine, you can see the results of your potential treatment in advance. What kind of lens would you like to wear? Choose the style of your liking...", color)
    var selection = npc.selection()

    if (plr.takeItem(5152013, 1)) {
        plr.setFace(color[selection])
        npc.sendBackNext("Tada~! Check it out!! What do you think? I really think your eyes look sooo fantastic now~~! Please come again ~", true, true)
    } else {
        npc.sendBackNext("I'm sorry, but I don't think you have our cosmetic lens coupon with you right now. Without the coupon, I'm afraid I can't do it for you..", true, true)
    }
}

// Generate by kimi-k2-instruct