// Main menu
var menu = "I'm the head of this hair salon Natalie. If you have #b#t5150053##k or #b#t5151036##k, allow me to take care of your hairdo. Please choose the one you want. \r\n#L0##bHaircut(VIP coupon)#l\r\n#L1#Dye your hair(VIP coupon)#l"
npc.sendSelection(menu)
var select = npc.selection()

if (select == 0) {
    // Haircut
    var hair
    if (plr.gender() < 1) {
        hair = [33040, 30060, 30210, 30140, 30200, 33170, 33100]
    } else {
        hair = [31150, 34090, 31300, 31700, 31350, 31740, 34110]
    }

    for (var i = 0; i < hair.length; i++) {
        hair[i] = hair[i] + (plr.hair() % 10)
    }

    npc.sendStyles("I can totally change up your hairstyle and make it look so good. Why don't you change it up a bit? with #b#t5150053##k I'll change it for you. Choose the one to your liking~", hair)
    var choice = npc.selection()

    if (plr.takeItem(5150053, 1)) {
        plr.setHair(hair[choice])
        npc.sendBackNext("Check it out!!. What do you think? Even I think this one is a work of art! AHAHAHA. Please let me know when you want another haircut, because I'll make you look good each time!", true, true)
    } else {
        npc.sendBackNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't give you a haircut without it. I'm sorry...", true, true)
    }
} else if (select == 1) {
    // Dye
    var base = Math.floor(plr.hair() / 10) * 10
    var hair = [base + 0, base + 1, base + 2, base + 4, base + 6]

    npc.sendStyles("I can totally change your haircolor and make it look so good. Why don't you change it up a bit? With #b#t5151036##k I'll change it for you. Choose the one to your liking.", hair)
    var choice = npc.selection()

    if (plr.takeItem(5151036, 1)) {
        plr.setHair(hair[choice])
        npc.sendBackNext("Check it out!!. What do you think? Even I think this one is a work of art! AHAHAHA. Please let me know when you want to dye your hair again, because I'll make you look good each time!", true, true)
    } else {
        npc.sendBackNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't dye your hair without it. I'm sorry...", true, true)
    }
}

// Generate by kimi-k2-instruct