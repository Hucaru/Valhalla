npc.sendBackNext("Hey, you look like you want to go farther and deeper past this place. Over there, though, you'll find yourself surrounded by aggressive, dangerous monsters, so even if you feel that you're ready to go, please be careful. Long ago, a few brave men from our town went in wanting to eliminate anyone threatening the town, but never came back out...", false, true)

if (plr.level() < 50) {
    npc.sendBackNext("If you are thinking of going in, I suggest you change your mind. But if you really want to go in...I'm only letting in the ones that are strong enough to stay alive in there. I do not wish to see anyone else die. Let's see ... Hmmm ... you haven't reached Level 50 yet. I can't let you in, then, so forget it.", true, false)
} else {
    if (npc.sendYesNo("If you are thinking of going in, I suggest you change your mind. But if you really want to go in...l'm only letting in the ones that are strong enough to stay alive in there. I do not wish to see anyone else die. Let's see ... Hmmm ...! You look pretty strong. All right, do you want go in?")) {
        plr.warp(211040300)
    } else {
        npc.sendBackNext("Even if your level's high it's hard to actually go in there, but if you ever change your mind please find me. After all, my job is to protect this place.", true, false)
    }
}

// Generate by kimi-k2-instruct