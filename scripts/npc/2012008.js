// Orbis Skin-Care service
npc.sendNext("Well, hello! Welcome to the Orbis Skin-Care~! Would you like to have a firm, tight, healthy looking skin like mine? With #b#t5153015##k, you can let us take care of the rest and have the kind of skin you've always wanted~!")

var skin = [0, 1, 2, 3, 4, 5, 9, 10, 11]
var sel = npc.sendAvatar("With our specialized machine, you can see yourself after the treatment in advance. What kind of skin-treatment would you like to do? Choose the style of your liking...", skin)

if (plr.itemCount(5153015) > 0) {
    plr.removeItemsByID(5153015, 1)
    plr.setSkinColor(skin[sel])
    npc.sendOk("Here's the mirror, check it out! Doesn't your skin look beautiful and glowing like mine? Hehe, it's wonderful. Please come again!")
} else {
    npc.sendNext("It looks like you don't have the coupon you need to receive the treatment. I'm sorry but it looks like we cannot do it for you.")
}