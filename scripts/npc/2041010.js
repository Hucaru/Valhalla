var couponFace = 5152006; // EXP Face Coupon

npc.sendNext("Hey there! If you have a #b#t" + couponFace + "##k, I can change your face for you!");

var z = plr.face() % 1000;
var baseMale = [20035, 20036, 20037, 20038, 20039];
var baseFemale = [21035, 21036, 21037, 21038, 21039];
var src = (plr.gender() < 1) ? baseMale : baseFemale;
var newFace = src[Math.floor(Math.random() * src.length)] + z;

if (plr.itemCount(couponFace) > 0) {
    plr.removeItemsByID(couponFace, 1);
    plr.setFace(newFace);
    npc.sendOk("All done! What do you think of your new look?");
} else {
    npc.sendOk("Hmm… looks like you don’t have a #b#t" + couponFace + "##k.");
}
