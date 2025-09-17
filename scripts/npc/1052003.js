npc.sendYesNo("I'll be on repair duty for a while. Do you have something to you need fixed?");
var answer = npc.selection();

if (answer) {
    npc.sendRepairWindow();
} else {
    npc.sendOk("Good items break easily. \r\nYou should repair them once in a while.");
}