if (plr.itemCount(4220137) > 0) {
    if (plr.getQuestStatus(20522) === 1 || plr.getQuestStatus(20526) === 1) {
        npc.sendOk("If it is that difficult for you to take care of Mimiana until it hatches out of the egg, then there's only one thing we can do. Neinheart may not like it, but walking is not that bad of an alternative. For now, you should just forfeit the #bRaising Mimiana#k Quest, and then talk to me so I can take away your egg.");
    } else {
        npc.sendOk("I know it's too bad, but it has come to this. If you wish to raise another one later, you're always welcome to see me.");
        plr.removeItemsByID(4220137, 1);
    }
} else {
    npc.sendOk("I don't understand what's you're saying.");
}