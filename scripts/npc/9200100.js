// Dr. Lenu — Cosmetic Lenses (Reg/VIP) in Henesys
const ITEM_REG = 5152010;  // Regular cosmetic lens coupon
const ITEM_VIP = 5152013;  // VIP cosmetic lens coupon

var faceId = plr.getFace();
var gender = plr.getGender(); // 0 = male, 1 = female

// greeting + menu
var sel = npc.sendMenu(
    "Why hello there! I'm Dr. Lenu, in charge of the cosmetic lenses here at the Henesys Plastic Surgery Shop! "
    + "With #b#t" + ITEM_REG + "##k or #b#t" + ITEM_VIP + "##k, you can have the kind of look you've always wanted! "
    + "All you have to do is find the cosmetic lens that most fits you, then let us take care of the rest. "
    + "Now, what would you like to use?",
    "#bCosmetic Lenses at Henesys (Reg coupon)#l",
    "#bCosmetic Lenses at Henesys (VIP coupon)#l"
);

function currentStyleBase() {
    // Keep the same eye shape, only changing its color block (hundreds)
    // teye = (face tail) + gender base (20000/21000)
    var teye = (faceId % 100) + (gender < 1 ? 20000 : 21000);
    return teye;
}

if (sel === 0) {
    // Regular coupon: random color within the same eye shape
    if (!npc.sendYesNo("If you use the regular coupon, you'll be awarded a random pair of cosmetic lenses. Are you going to use #b#t" + ITEM_REG + "##k and really make the change to your eyes?")) {
        npc.sendNext("I see... take your time and see if you really want it. Let me know when you've decided.");
    } else if (plr.itemCount(ITEM_REG) <= 0) {
        npc.sendNext("I'm sorry, but I don't think you have our cosmetic lens coupon with you right now. Without the coupon, I'm afraid I can't do it for you..");
    } else {
        var colors = [100, 200, 300, 400, 500, 600, 700];
        var teye = currentStyleBase();
        var newFace = teye + colors[Math.floor(Math.random() * colors.length)];

        plr.removeItemsByID(ITEM_REG, 1);
        plr.setFace(newFace);
        npc.sendNext("Tada~! Check it out!! What do you think? I really think your eyes look sooo fantastic now~~! Please come again ~");
    }
} else if (sel === 1) {
    // VIP coupon: preview specific colors for the current shape
    var colors = [100, 200, 300, 400, 500, 600, 700];
    var teye = currentStyleBase();
    var faces = [];
    for (var i = 0; i < colors.length; i++) {
        faces.push(teye + colors[i]);
    }

    var chosen = npc.sendAvatar.apply(
        npc,
        ["With our specialized machine, you can see the results of your potential treatment in advance. What kind of lens would you like to wear? Choose the style of your liking..."].concat(faces)
    );

    if (chosen < 0 || chosen >= faces.length) {
        npc.sendOk("Changed your mind? That's fine. Come back any time.");
    } else if (plr.itemCount(ITEM_VIP) <= 0) {
        npc.sendNext("I'm sorry, but I don't think you have our cosmetic lens coupon with you right now. Without the coupon, I'm afraid I can't do it for you..");
    } else {
        plr.removeItemsByID(ITEM_VIP, 1);
        plr.setFace(faces[chosen]);
        npc.sendNext("Tada~! Check it out!! What do you think? I really think your eyes look sooo fantastic now~~! Please come again ~");
    }
}