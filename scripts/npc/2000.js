// Quest ID constant
var Quest_Roger_Apple = 2

// Roger initial greeting
npc.sendSelection(
    "Hey! Nice weather today, huh?\r\n\r\n" +
    buildRogerMenu()
)

var sel = npc.selection()

if (plr.checkQuestStatus(Quest_Roger_Apple, 1) && plr.itemCount(2010000) >= 1) {
    npc.sendOk("You haven't eaten the #bApple#k that I gave you yet. Talk to me once you have!")
} else if (plr.checkQuestStatus(Quest_Roger_Apple, 1) && plr.itemCount(2010000) == 0) {
    npc.sendBackNext("Easy, right? You can set up a #bhotkey#k in the quickslots to the lower right of the screen to make it even easier. Oh, and your HP will automatically recover if you stand still, though it takes time.")
    npc.sendBackNext("Alright! I suppose after all that learning, you should receive a reward. This gift is a must for your travel in Maple World, so thank me! Use this for emergencies!\r\n\r\n#e#rREWARD:#k\r\n#b3 Apples\r\n+10 EXP#k", true, true)
    plr.giveItem(2010000, 3)
    plr.giveEXP(10)
    plr.completeQuest(Quest_Roger_Apple)
    plr.setQuestData(Quest_Roger_Apple, "end")
    npc.sendOk("Well, that's about all I can teach you. I know it's sad, but it is time to say goodbye. Take good care of yourself and do well, my friend!")
} else if (plr.checkQuestStatus(Quest_Roger_Apple, 0)) {
    npc.sendBackNext("Hey there, what's up! The name's Roger, and I'm here to teach you new, wide-eyed Maplers lots of cool things to help you get started.")
    npc.sendBackNext("You are asking who made me do this? No one! It's just all out of the overflowing kindness of my heart. Haha!", true, true)
    var ac = npc.sendYesNo("So... Let me just do this for fun! Abracadabra!")
    if (ac) {
        plr.startQuest(Quest_Roger_Apple)
        plr.setQuestData(Quest_Roger_Apple, "1")
        plr.giveHP(-25)
        plr.giveItem(2010000, 1)
        npc.sendBackNext("Ha! Your HP bar almost emptied! If your HP ever gets to 0, you're in trouble. To prevent that, consume food and potions. Here, take this #rRoger's Apple#k. Open your inventory (press #bI#k) and double-click the apple to eat it.")
        npc.sendOk("Eat the Roger's Apples to get back HP. Talk to me after you do.")
    } else {
        npc.sendOk("I can't believe I just got turned down!")
    }
}

function buildRogerMenu() {
    var status = plr.checkQuestStatus(Quest_Roger_Apple)
    if (status == 0) {
        return "#r#eQUEST AVAILABLE#k#n#l\r\n#L0##bRoger's Apple#k#l"
    } else if (status == 1) {
        if (plr.itemCount(2010000) >= 1) {
            return "#r#eQUEST IN PROGRESS#k#n#l\r\n#L0##bRoger's Apple (In Progress)#k#l"
        } else {
            return "#r#eQUEST THAT CAN BE COMPLETED#k#n#l\r\n#L0##bRoger's Apple (Ready to complete.)#k#l"
        }
    }
    return ""
}