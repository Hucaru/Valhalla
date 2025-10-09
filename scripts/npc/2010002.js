var couponFace = 5152005; // VIP Face Coupon

npc.sendSelection(
    "Welcome! If you have a #b#t" + couponFace + "##k, I can give you a brand new face!\r\n#L0#Get a face makeover (VIP coupon)#l"
);

if (npc.selection() === 0) {
    var z = plr.face() % 1000;
    var baseMale = [20020, 20021, 20022, 20023, 20024];
    var baseFemale = [21020, 21021, 21022, 21023, 21024];
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
