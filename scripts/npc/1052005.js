var couponFace = 5152000; // EXP Face Coupon

npc.sendNext("Hey there! If you have a #b#t" + couponFace + "##k, I can change your face for you!");

var z = plr.face() % 1000;
var baseMale = [20005, 20006, 20007, 20008, 20009];
var baseFemale = [21005, 21006, 21007, 21008, 21009];
var src = (plr.gender() < 1) ? baseMale : baseFemale;
var newFace = src[Math.floor(Math.random() * src.length)] + z;

if (plr.itemCount(couponFace) > 0) {
    plr.removeItemsByID(couponFace, 1);
    plr.setFace(newFace);
    npc.sendOk("All done! What do you think of your new look?");
} else {
    npc.sendOk("Hmm… looks like you don’t have a #b#t" + couponFace + "##k.");
}
