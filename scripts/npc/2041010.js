var face = []
if (plr.gender() < 1) {
    face = [20000, 20001, 20002, 20003, 20004, 20005, 20006, 20008, 20012, 20014, 20011]
} else {
    face = [21000, 21001, 21002, 21003, 21004, 21005, 21006, 21007, 21008, 21012, 21010]
}
for (var i = 0; i < face.length; i++) {
    face[i] = face[i] + Math.floor(plr.face() / 100 % 10) * 100
}

npc.sendStyles("Let's see... for #b#t5152057##k, you can get a new face. That's right. I can completely transform your face! Wanna give it a shot? Please consider your choice carefully.", face)
var selection = npc.selection()

if (plr.takeItem(5152057, 1)) {
    plr.setFace(face[selection])
    npc.sendOk("Ok, the surgery's over. See for it yourself.. What do you think? Quite fantastic, if I should say so myself. Please come again when you want another look, okay?")
} else {
    npc.sendOk("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...")
}

// Generate by kimi-k2-instruct