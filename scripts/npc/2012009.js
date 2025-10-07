var couponFace = 5152004; // EXP Face Coupon

npc.sendNext("Hey there! If you have a #b#t" + couponFace + "##k, I can change your face for you!");

var z = plr.face() % 1000;
var baseMale = [20025, 20026, 20027, 20028, 20029];
var baseFemale = [21025, 21026, 21027, 21028, 21029];
var src = (plr.gender() < 1) ? baseMale : baseFemale;
var newFace = src[Math.floor(Math.random() * src.length)] + z;

if (plr.itemCount(couponFace) > 0) {
    plr.removeItemsByID(couponFace, 1);
    plr.setFace(newFace);
    npc.sendOk("All done! What do you think of your new look?");
} else {
    npc.sendOk("Hmm… looks like you don’t have a #b#t" + couponFace + "##k.");
}
