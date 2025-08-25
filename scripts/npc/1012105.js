npc.sendBackNext("Welcome to Henesys Skin-Care! For just one teeny-weeny #b#t5153015##k, I can make your skin supple and glow-y, like mine! Trust me, you don't want to miss my facials.", false, true)

var skin = [0, 1, 2, 3, 4, 5, 9, 10, 11]
npc.sendStyles("We have the latest in beauty equipment. With our technology, you can preview what your skin will look like in advance! Which treatment would you like?", skin)

var selection = npc.selection()

if (plr.takeItem(5153015, 1)) {
    plr.setSkinColor(skin[selection])
    npc.sendBackNext("Here's the mirror, check it out! Doesn't your skin look beautiful and glowing like mine? Hehe, it's wonderful. Please come again!", true, true)
} else {
    npc.sendBackNext("It looks like you don't have the coupon you need to receive the treatment. I'm sorry but it looks like we cannot do it for you.", true, true)
}

// Generate by kimi-k2-instruct