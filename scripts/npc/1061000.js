// Repair system flow
if (npc.sendYesNo("I'll be on repair duty for a while. Do you have something to you need fixed?")) {
    npc.sendRepairWindow()
} else {
    npc.sendNext("Good items break easily. \r\nYou should repair them once in a while.")
}