// Pet Trainer Bartos
var infoAbout = "#b#m100000202##k: the road where you can take a walk with your pet, or train it through obstacles. If you aren't close with your pet yet, it may not follow your commands perfectly."

// Ask if player wants info or training
npc.sendSelection("Do you have any business with me?\r\n#L0#Please tell me about this place.#l")

if (npc.selection() == 0) {
    npc.sendOk(infoAbout)
}

// Check for letter
if (plr.itemCount(4031035) > 0) {
    npc.sendOk("Get that letter, jump over obstacles with your pet, and take that letter to my brother #p1012007#. Get him the letter and something good is going to happen to your pet.")
} else {
    if (npc.sendYesNo("This is the road where you can go take a walk with your pet. You can just walk around with it, or you can train your pet to go through the obstacles here. If you aren't too close with your pet yet, that may present a problem and he will not follow your command as much ... so, what do you think? Wanna train your pet?")) {
        if (plr.etcFreeSpace() < 1) {
            npc.sendOk("Your etc. inventory is full! I can't give you the letter unless there's room on ur inventory. Make an empty slot and then talk to me.")
        } else {
            plr.giveItem(4031035, 1)
            npc.sendOk("Ok, here's the letter. He wouldn't know I sent you if you just went there straight, so go through the obstacles with your pet, go to the very top, and then talk to #p1012007# to give him the letter. It won't be hard if you pay attention to your pet while going through obstacles. Good luck!")
        }
    } else {
        npc.sendOk("Hmmm ... too busy to do it right now? If you feel like doing it, though, come back and find me.")
    }
}