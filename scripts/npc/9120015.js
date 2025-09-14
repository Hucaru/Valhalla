npc.sendSelection("What do you want? \r\n#L0##bTell me about the hideout.#l\r\n#L1#Take me to the hideout.#l\r\n#L2#End conversation.#l");
var sel = npc.selection();

if (sel === 0) {
    npc.sendOk("I can take you to the hideout, but the place is crawling with thugs looking for trouble. Inside you'll also find the Yakuza Boss, who commands the underboss and all the lieutenants in the area. Getting into the hideout is the easy part, but you can only enter the room at the top floor ONCE a day. The boss's room is a no place to mess around. It's best you don't outstay your welcome. The big boss is a difficult foe, but the path to even reaching him will be filled with powerful enemies. You sure you can handle it?!");
} else if (sel === 1) {
    npc.sendBackNext("Oh, I've been waiting for you, hero. There's no turning back if we leave them alone. Before that happens, I would like you to use your power and teach the Yakuza Boss on the 5th floor a lesson. Don't let your guard down. Many strong people couldn't beat the Yakuza Boss, but l'm certain you can do it when I look into your eyes. Now go.", true, true);
    plr.warp(801040000);
} else { // sel === 2
    npc.sendOk("I don't have free time. Go back if you have no business.");
}