/*
** NPC: Denma the Owner
** Location: Henesys
** Purpose: Plastic Surgeon (VIP)
*/

const coupon = 5152001;

// Base faces; we'll add the player's current color block (hundreds)
var baseMale   = [20000, 20001, 20002, 20003, 20004, 20005, 20006, 20007, 20008];
var baseFemale = [21000, 21001, 21002, 21003, 21004, 21005, 21006, 21007, 21008];

// Compute color offset from current face
var currentFace = plr.face();
var colorOffset = (Math.floor(currentFace / 100) % 10) * 100;

// Gender-specific list with current color block applied
var faces = (plr.gender() < 1) ? baseMale.slice() : baseFemale.slice();
for (var i = 0; i < faces.length; i++) {
    faces[i] += colorOffset;
}

// Show preview selector (stateless flow: show, then read selection)
npc.sendAvatar.apply(
    npc,
    ["Let's see...for #b#t" + coupon + "##k, you can get a new face. That's right, I can completely transform your face! Wanna give it a shot? Please consider your choice carefully."]
        .concat(faces)
);

var sel = npc.selection();

// Validate and apply
if (sel < 0 || sel >= faces.length) {
    npc.sendOk("Changed your mind? That's fine. Come back any time.");
} else if (plr.itemCount(coupon) >= 1) {
    plr.removeItemsByID(coupon, 1);
    plr.setFace(faces[sel]);
    npc.sendOk("Alright, it's all done! Check yourself out in the mirror. Well, aren't you lookin' marvelous? Haha! If you're sick of it, just give me another call, alright?");
} else {
    npc.sendNext("Hmm ... it looks like you don't have the coupon specifically for this place. Sorry to say this, but without the coupon, there's no plastic surgery for you...");
}