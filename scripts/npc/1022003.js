if (npc.sendYesNo("I'll be on repair duty for a while. Do you have something to you need fixed?")) {
    npc.sendShop([[1000000, 0]]);   // repair window issued with empty shop
} else {
    npc.sendOk("Good items break easily. \r\nYou should repair them once in a while.");
}