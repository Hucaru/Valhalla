npc.sendSelection("I'm Mar the Fairy, and I can transfer stats from an existing pet to a new pet. \r\n#L0##bI want to transfer pet stats to a new pet.#l\r\n#L1#I want to revive a pet.#l")
var sel = npc.selection()

if (sel == 0) {
    npc.sendOk("I do not think you have the Pet AP Reset Scroll or a pet for closeness to be transferred with you... Cloy from henesys would definitely know about the Pet AP Reset Scroll...")
} else {
    npc.sendBackNext("I'm #p1032102#, and I'm researching all kinds of magic here in #m101000000#. I've been studying life magic for centuries, but there's always more to learn, it seems! Ah, excuse me as I get back to my research, then...", true, true)
}

// Generate by kimi-k2-instruct