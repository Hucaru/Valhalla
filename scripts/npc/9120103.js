npc.sendSelection("Hi, I pretty much shouldn't be doing this, but with a #b#t5152056##k or #b#t5152046##k, I will do it anyways for you. But don't forget, it will be random! Now, what would you like to use? \r\n#L0##bPlastic Surgery at Showa (REG coupon)#l\r\n#L1#Cosmetic Lenses at Showa (REG coupon)#l")
var select = npc.selection()

if (select == 0) {
    // Plastic Surgery
    var face
    if (plr.gender() < 1) {
        face = [20000, 20016, 20019, 20020, 20021, 20024, 20026]
    } else {
        face = [21000, 21002, 21009, 21016, 21022, 21025, 21027]
    }
    var randomFace = face[Math.floor(Math.random() * face.length)] + Math.floor(plr.face() / 100 % 10) * 100

    if (plr.takeItem(5152056, 1)) {
        plr.setFace(randomFace)
        npc.sendBackNext("Okay, the surgery's done. Here's a mirror--check it out. What a masterpiece, no? Haha! If you ever get tired of this look, please feel free to come visit me again.", true, true)
    } else {
        npc.sendBackNext("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...", true, true)
    }
} else if (select == 1) {
    // Cosmetic Lenses
    var color = [100, 200, 300, 400, 500, 600, 700]
    var teye = plr.face() % 100
    teye += plr.gender() < 1 ? 20000 : 21000
    var randomColor = color[Math.floor(Math.random() * color.length)] + teye

    if (plr.takeItem(5152046, 1)) {
        plr.setFace(randomColor)
        npc.sendBackNext("Here's the mirror. What do you think? I think they look tailor-made for you. I have to say, you look faaabulous. Please come again.", true, true)
    } else {
        npc.sendBackNext("I'm sorry, but I don't think you have our cosmetic lens coupon with you right now. Without the coupon, I'm afraid I can't do it for you..", true, true)
    }
}

// Generate by kimi-k2-instruct