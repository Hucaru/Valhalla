// Trainer Whoop conversation
if (plr.itemCount(4031128) > 0) {
    npc.sendOk("Get that letter, jump over obstacles with your pet, and take that letter to my brother #p2040033#. Get him the letter and something good is going to happen to your pet.")
} else {
    if (npc.sendYesNo("This is the road where you can go take a walk with your pet. You can just walk around with it, or you can train your pet to go through the obstacles here. If you aren't too close with your pet yet, that may present a problem and he will not follow your command as much... so, what do you think? Wanna train your pet?")) {
        npc.sendOk("Ok, here's the letter. He wouldn't know I sent you if you just went there straight, so go through the obstacles with your pet, go to the very top, and then talk to #p2040033# to give him the letter. It won't be hard if you pay attention to your pet while going through obstacles. Good luck!")
        plr.giveItem(4031128, 1)
    } else {
        npc.sendNext("Hmmm... too busy to do it right now? If you feel like doing it, though, come back and find me.")
    }
}