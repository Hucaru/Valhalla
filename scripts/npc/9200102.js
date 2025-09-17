// Dr. Bosch – Cosmetic Lenses

var regCoupon  = 5152012
var vipCoupon  = 5152015

npc.sendSelection(
    "Um... hi, I'm Dr. Bosch, and I am a cosmetic lens expert here at the Ludibrium Plastic Surgery Shop. I believe your eyes are the most important feature in your body, and with #b#t5152012##k or #b#t5152015##k, I can prescribe the right kind of cosmetic lenses for you. Now, what would you like to use?\r\n"
    + "#L0##bCosmetic Lenses at Ludibrium (Reg coupon)#l\r\n"
    + "#L1#Cosmetic Lenses at Ludibrium (VIP coupon)#l"
)
var sel = npc.selection()

if (sel === 0) {
    // Regular coupon – random colour
    if (npc.sendYesNo("If you use the regular coupon, I'll have to warn you that you'll be awarded a random pair of cosmetic lenses. Are you going to use #b#t5152012##k and really make the change to your eyes?")) {
        if (plr.itemCount(regCoupon) > 0) {
            plr.removeItemsByID(regCoupon, 1)

            var teye = plr.getFace() % 100
            teye     += plr.getGender() < 1 ? 20000 : 21000

            var colour = [100, 200, 300, 400, 500, 600, 700]
            var choice = colour[Math.floor(Math.random() * colour.length)] + teye

            plr.setFace(choice)
            npc.sendOk("Here's the mirror. What do you think? I think they look tailor-made for you. I have to say, you look faaabulous. Please come again.")
        } else {
            npc.sendOk("I'm sorry, but I don't think you have our cosmetic lens coupon with you right now. Without the coupon, I'm afraid I can't do it for you..")
        }
    }

} else if (sel === 1) {
    // VIP coupon – choose colour
    var teye = plr.getFace() % 100
    teye     += plr.getGender() < 1 ? 20000 : 21000

    var colours = []
    ;[100, 200, 300, 400, 500, 600, 700].forEach(c => colours.push(teye + c))

    var chosen = npc.sendAvatar("With our specialized machine, you can see yourself after the treatment in advance. What kind of lens would you like to wear? Choose the style of your liking...", ...colours)

    if (plr.itemCount(vipCoupon) > 0) {
        plr.removeItemsByID(vipCoupon, 1)
        plr.setFace(colours[chosen])
        npc.sendOk("Here's the mirror. What do you think? I think they look tailor-made for you. I have to say, you look faaabulous. Please come again.")
    } else {
        npc.sendOk("I'm sorry, but I don't think you have our cosmetic lens coupon with you right now. Without the coupon, I'm afraid I can't do it for you..")
    }
}