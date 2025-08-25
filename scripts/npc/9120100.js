// Welcome to Showa Hair-Salon
var menu = "Welcome, welcome, welcome to the showa Hair-Salon! Do you, by any chance, have #b#t5150053##k or #b#t5151036##k? If so, how about letting me take care of your hair? Please choose what you want to do with it. \r\n#L0##bChange hair style (VIP coupon)#l\r\n#L1#Dye your hair (VIP coupon)#l"
npc.sendSelection(menu)
var select = npc.selection()

if (select == 0) {
    // Change hairstyle
    var hair
    if (plr.gender() < 1) {
        hair = [30030, 33240, 30780, 30810, 30820, 30260, 30280, 30710, 30920, 30340]
    } else {
        hair = [31550, 31850, 31350, 31460, 31100, 31030, 31790, 31000, 31770, 34260]
    }
    for (var i = 0; i < hair.length; i++) {
        hair[i] = hair[i] + (plr.hair() % 10)
    }
    npc.sendStyles("I can change your hairstyle to something totally new. Aren't you sick of your current hairdo? With #b#t5150053##k, I can make that happen for you. Choose the hairstyle you'd like to sport.", hair)
    var styleChoice = npc.selection()
    if (plr.takeItem(5150053, 1)) {
        plr.setHair(hair[styleChoice])
        npc.sendBackNext("Enjoy your new and improved hairstyle!", true, true)
    } else {
        npc.sendBackNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't give you a haircut without it. I'm sorry...", true, true)
    }
} else if (select == 1) {
    // Dye hair
    var baseHair = Math.floor(plr.hair() / 10) * 10
    var hairColors = [baseHair + 0, baseHair + 1, baseHair + 2, baseHair + 3, baseHair + 4, baseHair + 5, baseHair + 6]
    npc.sendStyles("I can change your hair color to something totally new. Aren't you sick of your current hairdo? With #b#t5151036##k, I can make that happen. Choose the hair color you'd like to sport.", hairColors)
    var colorChoice = npc.selection()
    if (plr.takeItem(5151036, 1)) {
        plr.setHair(hairColors[colorChoice])
        npc.sendBackNext("Enjoy your new and improved hair colour!", true, true)
    } else {
        npc.sendBackNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't dye your hair without it. I'm sorry...", true, true)
    }
}

// Generate by kimi-k2-instruct