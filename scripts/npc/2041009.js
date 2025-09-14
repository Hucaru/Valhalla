// 咪尼 at Toy Town Beauty Salon
npc.sendSelection(
    "Hi, I'm the assistant here. Don't worry, I'm plenty good enough for this. If you have #b#t5150052##k or #b#t5151035##k by any chance, then allow me to take care of the rest, alright? \r\n" +
    "#L0##bChange hair-style (Regular coupon)#l\r\n" +
    "#L1#Dye your hair (Regular coupon)#l"
)

var select = npc.selection()
var hair

if (select === 0) {
    // -> action0: Change hair-style
    if (plr.job() < 1000) {                 // < 1 -> male
        hair = [30250, 30190, 30150, 30050, 30280, 30240, 30300, 30160, 30650, 30540, 30640, 30680]
    } else {
        hair = [31150, 31280, 31160, 31120, 31290, 31270, 31030, 31230, 31010, 31640, 31540, 31680, 31600]
    }
    hair = hair[Math.floor(Math.random() * hair.length)] + (plr.getHair() % 10)

    if (npc.sendYesNo("If you use the regular coupon, your hair-style will be changed into a random new look. Are you sure you want to use #b#t5150052##k and change it?")) {
        if (plr.itemCount(5150052) > 0) {
            plr.giveItem(5150052, -1)
            plr.setHair(hair)
            npc.sendOk("Hey, here's the mirror. What do you think of your new haircut? I know it wasn't the smoothest of all, but didn't it come out pretty nice? If you ever feel like changing it up again later, please drop by.")
        } else {
            npc.sendOk("Hmmm...are you sure you have our designated coupon? Sorry but no haircut without it.")
        }
    }

} else if (select === 1) {
    // -> action1: Dye hair
    hair = Math.floor(plr.getHair() / 10) * 10
    var colors = [hair + 0, hair + 1, hair + 2, hair + 3, hair + 4, hair + 5]
    hair = colors[Math.floor(Math.random() * colors.length)]

    if (npc.sendYesNo("If you use the regular coupon, your hair-color will be changed into a random new look. Are you sure you want to use #b#t5151035##k and change it?")) {
        if (plr.itemCount(5151035) > 0) {
            plr.giveItem(5151035, -1)
            plr.setHair(hair)
            npc.sendOk("Hey, here's the mirror. What do you think of your new haircolor? I know it wasn't the smoothest of all, but didn't it come out pretty nice? If you ever feel like changing it up again later, please drop by.")
        } else {
            npc.sendOk("Hmmm...are you sure you have our designated coupon? Sorry but no dye your hair without it.")
        }
    }
}