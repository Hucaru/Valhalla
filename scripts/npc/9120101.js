// Assistant greeting
var menu = "Hi, I'm the assistant here. If you have #b#t5150052##k or #b#t5151035##k, please allow me to change your hairdo..\r\n"
menu += "#L0##bChange Hairstyle (REG Coupon)#l\r\n"
menu += "#L1#Dye Your Hair (REG Coupon)#l"
npc.sendSelection(menu)
var select = npc.selection()

if (select == 0) {
    // Change hairstyle
    var hair
    if (plr.gender() < 1) {
        hair = [30000, 30120, 30140, 30190, 30210, 30360, 30220, 30370, 30400, 30440, 30790, 30800, 30810, 30770, 30760]
    } else {
        hair = [31030, 31050, 31000, 31070, 31100, 31120, 31130, 31250, 31340, 31680, 31350, 31400, 31650, 31550, 31800]
    }
    var chosen = hair[Math.floor(Math.random() * hair.length)] + (plr.hair() % 10)

    if (npc.sendYesNo("If you use the REG coupon your hair will change to a RANDOM new hairstyle. Would you like to use #b#t5150052##k to change your hairstyle?")) {
        if (plr.takeItem(5150052, 1)) {
            plr.setHair(chosen)
            npc.sendBackNext("Now, here's the mirror. What do you think of your new haircut? Doesn't it look nice for a job done by an assistant? Come back later when you need to change it up again!", true, true)
        } else {
            npc.sendBackNext("Hmmm...are you sure you have our designated coupon? Sorry but no haircut without it.", true, true)
        }
    }
} else if (select == 1) {
    // Dye hair
    var base = Math.floor(plr.hair() / 10) * 10
    var colors = [base + 0, base + 1, base + 2, base + 3, base + 4, base + 5]
    var chosen = colors[Math.floor(Math.random() * colors.length)]

    if (npc.sendYesNo("If you use the REG coupon your hair will change to a RANDOM new haircolor. Would you like to use #b#t5151035##k to change your haircolor?")) {
        if (plr.takeItem(5151035, 1)) {
            plr.setHair(chosen)
            npc.sendBackNext("Now, here's the mirror. What do you think of your new haircolor? Doesn't it look nice for a job done by an assistant? Come back later when you need to change it up again!", true, true)
        } else {
            npc.sendBackNext("Hmmm...are you sure you have our designated coupon? Sorry but no dye your hair without it.", true, true)
        }
    }
}

// Generate by kimi-k2-instruct