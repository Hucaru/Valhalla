```javascript
var itemId = 5152057;
var face = [];

if (plr.job() < 1000) {   // male
    face = [20000, 20001, 20002, 20003, 20004, 20005, 20006, 20008, 20012, 20014, 20011];
} else {                  // female
    face = [21000, 21001, 21002, 21003, 21004, 21005, 21006, 21007, 21008, 21012, 21010];
}

for (var i = 0; i < face.length; i++)
    face[i] += Math.floor(plr.face / 100) % 10 * 100;

if (plr.itemCount(itemId)) {
    var sel = npc.askAvatar("Let's see... for #b#t5152057##k, you can get a new face. That's right. I can completely transform your face! Wanna give it a shot? Please consider your choice carefully.", face);
    
    plr.removeItemsByID(itemId, 1);
    plr.setFace(face[sel]);
    npc.sendBackNext("Ok, the surgery's over. See for it yourself.. What do you think? Quite fantastic, if I should say so myself. Please come again when you want another look, okay?", false, true);
} else {
    npc.sendOk("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...");
}
```