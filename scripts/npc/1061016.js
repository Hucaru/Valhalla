npc.sendSelection("You again? I sure see you a lot. What do you want? \r\n\r\n#L0##bPlease make Leather Shoes with 20 pieces of Balrog Leather.#l")
var sel = npc.selection()

if (plr.itemCount(4001261) < 20) {
    npc.sendOk("This isn't enough Balrog Leather to make anything. Shoes just aren't going to happen.")
} else if (plr.giveItem(1072375, 1)) {
    plr.removeItemsByID(4001261, 20)
    npc.sendOk("What do you think? Not a bad pair of shoes right? They may look plain, but that Balrog Leather makes them tough as nails!")
} else {
    npc.sendOk("See if you have enough material or if you have enough space available in Equip.")
}