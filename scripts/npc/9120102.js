npc.sendSelection("I'm in charge of the Plastic Surgery here at Showa Shop! I believe your eyes are the most important feature in your body, and with #b#t5152057##k or #b#t5152045##k, I can prescribe the right kind of plastic surgery and cosmetic lenses for you. Now, what would you like to use? \r\n#L0##bPlastic Surgery at Showa (VIP coupon)#l\r\n#L1#Cosmetic Lenses at Showa (VIP coupon)#l")
var select = npc.selection()

if (select == 0) {
    var face
    if (plr.gender() < 1) {
        face = [20020, 20000, 20002, 20004, 20005, 20012]
    } else {
        face = [21021, 21000, 21002, 21003, 21006, 21012]
    }
    for (var i = 0; i < face.length; i++) {
        face[i] = face[i] + parseInt(plr.face() / 100 % 10) * 100
    }
    npc.sendStyles("Let's see... for #b#t5152057##k, you can get a new face. That's right, I can completely transform your face! Wanna give it a shot? Please consider your choice carefully.", face)
    var sel = npc.selection()
    if (plr.takeItem(5152057, 1)) {
        plr.setFace(face[sel])
        npc.sendBackNext("Alright, it's all done! Check yourself out in the mirror. Well, aren't you lookin' marvelous? Haha! If you're sick of it, just give me another call, alright?", true, true)
    } else {
        npc.sendBackNext("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...", true, true)
    }
} else if (select == 1) {
    var color = [100, 200, 300, 400, 500, 600, 700]
    var teye = plr.face() % 100
    teye += plr.gender() < 1 ? 20000 : 21000
    for (var i = 0; i < color.length; i++) {
        color[i] = teye + color[i]
    }
    npc.sendStyles("With our specialized machine, you can see the results of your potential treatment in advance. What kind of lens would you like to wear? Choose the style of your liking...", color)
    var sel = npc.selection()
    if (plr.takeItem(5152045, 1)) {
        plr.setFace(color[sel])
        npc.sendBackNext("Enjoy your new and improved cosmetic lenses!", true, true)
    } else {
        npc.sendBackNext("I'm sorry, but I don't think you have our cosmetic lens coupon with you right now. Without the coupon, I'm afraid I can't do it for you..", true, true)
    }
}

// Generate by kimi-k2-instruct