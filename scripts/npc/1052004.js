var couponFace = 5152001; // VIP Face Coupon

npc.sendSelection(
    "Welcome! If you have a #b#t" + couponFace + "##k, I can give you a brand new face!\r\n#L0#Get a face makeover (VIP coupon)#l"
);

if (npc.selection() === 0) {
    var z = plr.face() % 1000;
    var baseMale = [20000, 20001, 20002, 20003, 20004];
    var baseFemale = [21000, 21001, 21002, 21003, 21004];
    var src = (plr.gender() < 1) ? baseMale : baseFemale;
    var faceList = [];
    for (var i = 0; i < src.length; i++) {
        faceList.push(src[i] + z);
    }
    npc.sendAvatar.apply(npc, ["Choose the face you want!"].concat(faceList));
    var choice = npc.selection();
    if (choice < 0 || choice >= faceList.length) {
        npc.sendOk("Changed your mind? That’s fine, come back any time.");
    } else if (plr.itemCount(couponFace) > 0) {
        plr.removeItemsByID(couponFace, 1);
        plr.setFace(faceList[choice]);
        npc.sendOk("Enjoy your new look!");
    } else {
        npc.sendOk("It seems like you don’t have a #b#t" + couponFace + "##k.");
    }
}
