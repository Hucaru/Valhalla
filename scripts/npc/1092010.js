npc.sendSelection("You're trying to discard the Dirty Treasure Map or Eggs? Well, you've come to the right place. \r\n\r\n#b#L0#I want to discard the Dirty Treasure Map.#l\r\n#L1#I want to discard the Eggs.#l");
var sel = npc.selection();

if (sel === 0) {
    if (plr.itemCount(4032094) >= 1) {
        plr.removeItemsByID(4032094, 1);
        npc.sendOk("Dirty Treasure Map discarded.");
    } else {
        npc.sendOk("Hmm... What is it? I'm the person that discards the Dirty Treasure Maps, but you don't seem to have the Dirty Treasure Map.");
    }
} else if (sel === 1) {
    if (plr.itemCount(4032095) >= 1) {
        plr.removeItemsByID(4032095, 1);
        npc.sendOk("Eggs discarded.");
    } else {
        npc.sendOk("Hmm... What is it? I'm the person that discards the Eggs, but you don't seem to have the Egg.");
    }
}