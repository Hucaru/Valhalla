npc.sendSelection("Hi, I'm the assistant here. If you have #b#t5150052##k or #b#t5151035##k, please allow me to change your hairdo..\r\n#L0##bChange Hairstyle (REG Coupon)#l\r\n#L1#Dye Your Hair (REG Coupon)#l");
var sel = npc.selection();

if (sel === 0) {
    var hair;
    if (plr.getGender() < 1) {
        hair = [30000, 30120, 30140, 30190, 30210, 30360, 30220, 30370, 30400, 30440, 30790, 30800, 30810, 30770, 30760];
    } else {
        hair = [31030, 31050, 31000, 31070, 31100, 31120, 31130, 31250, 31340, 31680, 31350, 31400, 31650, 31550, 31800];
    }
    hair = hair[Math.floor(Math.random() * hair.length)] + (plr.getHair() % 10);

    if (npc.sendYesNo("If you use the REG coupon your hair will change to a RANDOM new hairstyle. Would you like to use #b#t5150052##k to change your hairstyle?")) {
        if (plr.itemCount(5150052) > 0) {
            plr.giveItem(5150052, -1);
            plr.setHair(hair);
            npc.sendOk("Now, here's the mirror. What do you think of your new haircut? Doesn't it look nice for a job done by an assistant? Come back later when you need to change it up again!");
        } else {
            npc.sendOk("Hmmm...are you sure you have our designated coupon? Sorry but no haircut without it.");
        }
    }
} else if (sel === 1) {
    var base = Math.floor(plr.getHair() / 10) * 10;
    var hairList = [base + 0, base + 1, base + 2, base + 3, base + 4, base + 5];
    var hair = hairList[Math.floor(Math.random() * hairList.length)];

    if (npc.sendYesNo("If you use the REG coupon your hair will change to a RANDOM new haircolor. Would you like to use #b#t5151035##k to change your haircolor?")) {
        if (plr.itemCount(5151035) > 0) {
            plr.giveItem(5151035, -1);
            plr.setHair(hair);
            npc.sendOk("Now, here's the mirror. What do you think of your new haircolor? Doesn't it look nice for a job done by an assistant? Come back later when you need to change it up again!");
        } else {
            npc.sendOk("Hmmm...are you sure you have our designated coupon? Sorry but no dye your hair without it.");
        }
    }
}