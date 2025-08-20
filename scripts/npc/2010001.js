npc.sendSelection("Hello I'm Mino the Owner. If you have either #b#t5150053##k or #b#t5151036##k, then please let me take care of your hair. Choose what you want to do with it. \r\n#L0##bHaircut(VIP coupon)#l\r\n#L1#Dye your hair(VIP coupon)#l")
var select = npc.selection()

if (select == 0) {
    // Haircut
    var hair
    if (plr.gender() < 1) {
        hair = [33240, 30230, 30490, 30260, 30280, 33050, 30340]
    } else {
        hair = [34060, 31220, 31110, 31790, 31230, 31630, 34260]
    }
    for (var i = 0; i < hair.length; i++) {
        hair[i] = hair[i] + (plr.hair() % 10)
    }
    npc.sendStyles("I can completely change the look of your hair. Are you ready for a change by any chance? With #b#t5150053##k I can change it up for you. please choose the one you want!", hair)
    var choice = npc.selection()
    if (plr.takeItem(5150053, 1)) {
        plr.setHair(hair[choice])
        npc.sendBackNext("Enjoy your new and improved hairstyle!", true, true)
    } else {
        npc.sendBackNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't give you a haircut without it. I'm sorry...", true, true)
    }
} else if (select == 1) {
    // Dye
    var base = Math.floor(plr.hair() / 10) * 10
    var hair = [base + 0, base + 1, base + 3, base + 4, base + 5]
    npc.sendStyles("I can completely change your haircolor. Are you ready for a change by any chance? With #b#t5151036##k I can change it up for you. Please choose the one you want!", hair)
    var choice = npc.selection()
    if (plr.takeItem(5151036, 1)) {
        plr.setHair(hair[choice])
        npc.sendBackNext("Enjoy your new and improved hair colour!", true, true)
    } else {
        npc.sendBackNext("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't dye your hair without it. I'm sorry...", true, true)
    }
}

// Generate by kimi-k2-instruct