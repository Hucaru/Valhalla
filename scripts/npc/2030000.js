npc.sendNext("Hey, you look like you want to go farther and deeper past this place. Over there, though, you'll find yourself surrounded by aggressive, dangerous monsters, so even if you feel that you're ready to go, please be careful. Long ago, a few brave men from our town went in wanting to eliminate anyone threatening the town, but never came back out...")

if (plr.getLevel() < 50) {
    npc.sendBack("If you are thinking of going in, I suggest you change your mind. But if you really want to go in...I'm only letting in the ones that are strong enough to stay alive in there. I do not wish to see anyone else die. Let's see ... Hmmm ... you haven't reached Level 50 yet. I can't let you in, then, so forget it.")
} else {
    if (npc.sendYesNo("If you are thinking of going in, I suggest you change your mind. But if you really want to go in...l'm only letting in the ones that are strong enough to stay alive in there. I do not wish to see anyone else die. Let's see ... Hmmm ...! You look pretty strong. All right, do you want go in?")) {
        npc.sendBackNext("Okay, I got it! Just let us do the work, and you'll get there in the blink of an eye! Oh, and this won't cost you any money.")
        plr.warp(211040300)
    } else {
        npc.sendBack("Even if your level's high it's hard to actually go in there, but if you ever change your mind please find me. After all, my job is to protect this place.")
    }
}