npc.sendYesNo("If you use the regular coupon, you may end up with a random new look for your face...do you still want to do it using #b#t5152056##k?")

var face
if (plr.job() < 1) {
    face = [20000, 20005, 20008, 20012, 20016, 20022, 20032]
} else {
    face = [21000, 21002, 21008, 21014, 21020, 21024, 21029]
}

var chosen = face[Math.floor(Math.random() * face.length)] + Math.floor(plr.face() / 100 % 10) * 100

if (plr.takeItem(5152056, 1)) {
    plr.setFace(chosen)
    npc.sendBackNext("Okay, the surgery's done. Here's a mirror--check it out. What a masterpiece, no? Haha! If you ever get tired of this look, please feel free to come visit me again.", true, true)
} else {
    npc.sendBackNext("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...", true, true)
}

// Generate by kimi-k2-instruct