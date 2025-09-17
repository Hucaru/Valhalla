// Welcome message
npc.sendSelection("Welcome, welcome, welcome to the showa Hair-Salon! Do you, by any chance, have #b#t5150053##k or #b#t5151036##k? If so, how about letting me take care of your hair? Please choose what you want to do with it.\r\n#L0##bChange hair style (VIP coupon)#l\r\n#L1##bDye your hair (VIP coupon)#l");
var select = npc.selection();

if (select === 0) {
  let hair;
  if (plr.job() < 1000) {
    hair = [30030, 33240, 30780, 30810, 30820, 30260, 30280, 30710, 30920, 30340];
  } else {
    hair = [31550, 31850, 31350, 31460, 31100, 31030, 31790, 31000, 31770, 34260];
  }

  for (let i = 0; i < hair.length; i++) {
    hair[i] = hair[i] + (plr.getHair() % 10);
  }

  var choice = npc.askAvatar("I can change your hairstyle to something totally new. Aren't you sick of your current hairdo? With #b#t5150053##k, I can make that happen for you. Choose the hairstyle you'd like to sport.", hair);

  if (plr.itemCount(5150053) > 0) {
    plr.giveItem(5150053, -1);
    plr.setHair(hair[choice]);
    npc.sendOk("Enjoy your new and improved hairstyle!");
  } else {
    npc.sendOk("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't give you a haircut without it. I'm sorry...");
  }

} else if (select === 1) {
  let base = Math.floor(plr.getHair() / 10) * 10;
  let hairColor = [base + 0, base + 1, base + 2, base + 3, base + 4, base + 5, base + 6];

  var choice = npc.askAvatar("I can change your hair color to something totally new. Aren't you sick of your current hairdo? With #b#t5151036##k, I can make that happen. Choose the hair color you'd like to sport.", hairColor);

  if (plr.itemCount(5151036) > 0) {
    plr.giveItem(5151036, -1);
    plr.setHair(hairColor[choice]);
    npc.sendOk("Enjoy your new and improved hair colour!");
  } else {
    npc.sendOk("Hmmm...it looks like you don't have our designated coupon...I'm afraid I can't dye your hair without it. I'm sorry...");
  }
}