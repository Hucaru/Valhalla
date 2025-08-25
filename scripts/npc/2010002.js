var face = []
if (plr.gender() < 1) {
    face = [20000, 20001, 20002, 20003, 20004, 20005, 20006, 20008, 20012, 20014, 20022, 20028]
} else {
    face = [21000, 21001, 21002, 21003, 21004, 21005, 21006, 21007, 21008, 21012, 21023, 21026]
}

for (var i = 0; i < face.length; i++) {
    face[i] = face[i] + Math.floor(plr.face() / 100 % 10) * 100
}

var selection = npc.sendStyles("Welcome, welcome! Not happy with your look? Neither am I. But for #b#t5152057##k, I can transform your face and get you the look you've always wanted.", face)

if (plr.takeItem(5152057, 1)) {
    plr.setFace(face[selection])
    npc.sendOk("Ok, the surgery's over. See for it yourself.. What do you think? Quite fantastic, if I should say so myself. Please come again when you want another look, okay?")
} else {
    npc.sendOk("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...")
}

// Generate by kimi-k2-instruct