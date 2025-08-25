npc.sendBackNext("Oh, hello! Welcome to the Ludibrium Skin-Care! Are you interested in getting tanned and looking sexy? How about a beautiful, snow-white skin? If you have #b#t5153015##k, you can let us take care of the rest and have the kind of skin you've always dreamed of!", false, true)

var skin = [0, 1, 2, 3, 4, 5, 9, 10, 11]
npc.sendStyles("With our specialized machine, you can see yourself after the treatment in advance. What kind of skin-treatment would you like to do? Choose the style of your liking...", skin)
var selection = npc.selection()

if (plr.takeItem(5153015, 1)) {
    plr.setSkinColor(skin[selection])
    npc.sendBackNext("Here's the mirror, check it out! Doesn't your skin look beautiful and glowing like mine? Hehe, it's wonderful. Please come again!", true, true)
} else {
    npc.sendBackNext("It looks like you don't have the coupon you need to receive the treatment. I'm sorry but it looks like we cannot do it for you.", true, true)
}

// Generate by kimi-k2-instruct