npc.sendYesNo("I'll be on repair duty for a while. Do you have something you need fixed?")
if (npc.selection() === 1) {
    npc.sendRepair()
} else {
    npc.sendOk("Good items break easily. \r\nYou should repair them once in a while.")
}