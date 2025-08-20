// Persian Cat quest flow
if (plr.getQuestStatus(8012) != 1) {
    npc.sendOk("Haha... you dare attempt to anwer my wickedly hard questions? Well, they aren't free--but the prize is worth it!")
    return
}

if (plr.hasItem(4031064, 1)) {
    npc.sendOk("Meeeoooowww!")
    return
}

if (plr.getFreeSlots(4) < 1) {
    npc.sendOk("Please make sure you have at least 1 empty slot in your Etc tab.")
    return
}

if (!npc.sendYesNo("Did you get them all? Are you going to try to answer all of my questions?")) {
    npc.sendBackNext("You don't have the courage to face these questions. I knew it...out of my sight!", true, true)
    return
}

if (plr.itemQuantity(2020001) < 300) {
    npc.sendBackNext("Hey, are you sure you brought the 300 Fried Chickens I asked for? Check again and see if you brought enough.", true, true)
    return
}

plr.takeItem(2020001, 300)
npc.sendBackNext("Good job! The alley cats are gonna feast tonight! Now, on to my questions. I'm sure you're aware of this, but remember, if you get a single one wrong, it's over. This is all or nothing!", true, true)

// Question 1
var q1 = "Question no.1: What's the name of the vegetable store owner in Showa Town? \r\n#L0#Sami#l\r\n#L1#Kami#l\r\n#L2#Umi#l"
npc.sendSelection(q1)
var sel1 = npc.selection()
if (sel1 != 2) {
    npc.sendOk("Hmmm...all humans make mistakes! If you want to take another crack at it, then bring me 300 Fried Chickens.")
    return
}

// Question 2
var q2 = "Question no.2: Which of these NPCs does NOT stand in front of the movie theater at Showa Town? \r\n#L0#Sky#l\r\n#L1#Furano#l\r\n#L2#Shinta#l"
npc.sendSelection(q2)
var sel2 = npc.selection()
if (sel2 != 2) {
    npc.sendOk("Hmmm...all humans make mistakes! If you want to take another crack at it, then bring me 300 Fried Chickens.")
    return
}

// Question 3
var q3 = "Question no.3: What is the name of NPC that transfers travelers from Showa Town to the Mushroom Shrine? \r\n#L0#Perry#l\r\n#L1#Spinel#l\r\n#L2#Transporter#l"
npc.sendSelection(q3)
var sel3 = npc.selection()
if (sel3 != 0) {
    npc.sendOk("Hmmm...all humans make mistakes! If you want to take another crack at it, then bring me 300 Fried Chickens.")
    return
}

npc.sendBackNext("Wow, you answered all the questions correctly! I may not be the most fond of humans, but I HATE breaking a promise! So, as promised, here's the Orange Marble. You earned it!", true, true)
plr.giveItem(4031064, 1)
npc.sendOk("Our business is concluded, thank you very much! You can leave now!")

// Generate by kimi-k2-instruct