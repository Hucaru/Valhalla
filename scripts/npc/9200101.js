npc.sendSelection("Why hello there! I'm Dr. Lenu... \r\n#L0##bCosmetic Lenses at Henesys (Reg coupon)#l\r\n#L1#Cosmetic Lenses at Henesys (VIP coupon)#l");
var sel = npc.selection();

if (sel === 0) {
    if (!npc.sendYesNo("If you use the regular coupon, you'll be awarded a random pair of cosmetic lenses. Are you going to use #b#t5152010##k and really make the change to your eyes?")) {
        // user said no
        npc.send("Maybe next time!");
    }

    if (plr.itemCount(5152011) > 0) {
        // consume one 5152011
        if (!plr.removeItemsByID(5152011, 1)) {
            npc.sendOk("Huh? I couldn't find the coupon in your inventory.");
        }
        // Use avatar window with random among 7 colors (simulate random: pick a prebuilt set and choose)
        var options = [100,200,300,400,500,600,700].map(function(v){ return v; });
        var pick = Math.floor(Math.random() * options.length);
        // Show a confirmation image/text for flavor
        npc.send("Here's the mirror. What do you think? Fabulous!");
    } else {
        npc.send("I'm sorry, but I don't think you have our cosmetic lens coupon with you right now. Without the coupon, I'm afraid I can't do it for you.");
    }
} else if (sel === 1) {
    // VIP: let the player pick with sendAvatar
    // Build the styles you want to show (IDs must be valid style IDs)
    var styles = [21000+100,21000+200,21000+300,21000+400,21000+500,21000+600,21000+700];
    var pick = npc.sendAvatar("With our specialized machine, you can preview your new look. Which lens would you like?", styles);

    if (plr.itemCount(5152014) > 0) {
        if (!plr.removeItemsByID(5152014, 1)) {
            npc.sendOk("Huh? I couldn't consume the coupon.");
        }
        // Apply the chosen style via the style UI effect (actual stat change requires face setter in API)
        npc.send("Here's the mirror. Looking good! Please come again.");
    } else {
        npc.send("I'm sorry, but I don't think you have our VIP lens coupon with you right now.");
    }
}