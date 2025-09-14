// Early “no quest in progress” stop  
if (!plr.checkQuestStatus(8012, 1)) {
    npc.sendOk("Haha... you dare attempt to anwer my wickedly hard questions? Well, they aren't free--but the prize is worth it!")
}

// Already has the marble  
if (plr.itemCount(4031064) > 0) {
    npc.sendOk("Meeeoooowww!")
}

// Not enough inventory space  
if (plr.getInventoryFree(4) < 1) {        // 4 == MapleInventoryType.ETC
    npc.sendOk("Please make sure you have at least 1 empty slot in your Etc tab.")
}

npc.sendBackNext("Did you get them all? Are you going to try to answer all of my questions?", false, true)

// Cost confirmation  
if (plr.itemCount(2020001) < 300) {
    npc.sendOk("Hey, are you sure you brought the 300 Fried Chickens I asked for? Check again and see if you brought enough.")
}

// Question 1  
var sel1 = npc.sendMenu("Good job! The alley cats are gonna feast tonight! Now, on to my questions. I'm sure you're aware of this, but remember, if you get a single one wrong, it's over. This is all or nothing!\r\n#b\r\n#L0#Sami\r\n#L1#Kami\r\n#L2#Umi")
if (sel1 !== 2) {
    npc.sendOk("Hmmm...all humans make mistakes! If you want to take another crack at it, then bring me 300 Fried Chickens.")
}

// Question 2  
var sel2 = npc.sendMenu("Question no.2: Which of these NPCs does NOT stand in front of the movie theater at Showa Town?\r\n#b\r\n#L0#Sky\r\n#L1#Furano\r\n#L2#Shinta")
if (sel2 !== 2) {
    npc.sendOk("Hmmm...all humans make mistakes! If you want to take another crack at it, then bring me 300 Fried Chickens.")
}

// Question 3  
var sel3 = npc.sendMenu("Question no.3: What is the name of NPC that transfers travelers from Showa Town to the Mushroom Shrine?\r\n#b\r\n#L0#Perry\r\n#L1#Spinel\r\n#L2#Transporter")
if (sel3 !== 0) {
    npc.sendOk("Hmmm...all humans make mistakes! If you want to take another crack at it, then bring me 300 Fried Chickens.")
}

// Success  
npc.sendBackNext("Wow, you answered all the questions correctly! I may not be the most fond of humans, but I HATE breaking a promise! So, as promised, here's the Orange Marble. You earned it!", true, true)

// Remove chickens & give marble  
plr.removeItemsByID(2020001, 300)
plr.giveItem(4031064, 1)

npc.sendOk("Our business is concluded, thank you very much! You can leave now!")