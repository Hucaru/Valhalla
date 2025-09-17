// npc is Andres (Andre) at 103000005
const codeIter = () => {
    npc.sendSelection("I'm Andres, Don's assistant. Everyone calls me Andre, though. If you have #b#t5150052##k or #b#t5151035##k, please let me change your hairdo ... \r\n#L0##bHaircut(REG coupon)#l\r\n#L1#Dye your hair(REG coupon)#l");
    const select = npc.selection();

    // ====== action0 – haircut ======
    if (select === 0) {
        // pick random cut
        let hairOptions;
        if (plr.job() !== 1) {           // male check (0 = male, 1 = female)
            hairOptions = [30130, 30350, 30190, 30110, 30180, 30050, 30040, 30160, 30770, 30620, 30550, 30520];
        } else {
            hairOptions = [31060, 31090, 31020, 31130, 31120, 31140, 31330, 31010, 31520, 31440, 31750, 31620];
        }
        let chosen = hairOptions[Math.floor(Math.random() * hairOptions.length)] + (plr.hair() % 10);

        if (npc.sendYesNo("If you use the REG coupon your hair will change RANDOMLY with a chance to obtain a new experimental style that I came up with. Are you going to use #b#t5150052##k and really change your hairstyle?")) {
            if (plr.itemCount(5150052) > 0) {
                plr.removeItemsByID(5150052, 1);
                plr.setHair(chosen);
                npc.sendBackNext("Ok, here's the mirror. Your new haircut! What do you think? I know it wasn't the smoothest, but it still looks pretty good! Come back later when you need to change it up again!", false, false);
            } else {
                npc.sendBackNext("Hmmm...are you sure you have our designated coupon? Sorry but no haircut without it.", false, false);
            }
        } else {
            npc.sendBackNext("I see...think about it a little more and if you want to do it, come talk to me.", false, false);
        }

    // ====== action1 – dye ======
    } else if (select === 1) {
        // pick random color
        let base = Math.floor(plr.hair() / 10) * 10;
        let dyeOptions = [base, base + 1, base + 2, base + 3, base + 4, base + 5];
        let chosenColor = dyeOptions[Math.floor(Math.random() * dyeOptions.length)];

        if (npc.sendYesNo("If you use a regular coupon your hair will change RANDOMLY. Do you still want to use #b#t5151035##k and change it up?")) {
            if (plr.itemCount(5151035) > 0) {
                plr.removeItemsByID(5151035, 1);
                plr.setHair(chosenColor);
                npc.sendBackNext("Ok, here's the mirror. Your new haircolor! What do you think? I know it wasn't the smoothest, but it still looks pretty good! Come back later when you need to change it up again!", false, false);
            } else {
                npc.sendBackNext("Hmmm...are you sure you have our designated coupon? Sorry but no dye your hair without it.", false, false);
            }
        } else {
            npc.sendBackNext("I see...think about it a little more and if you want to do it, come talk to me.", false, false);
        }
    }
};
codeIter();