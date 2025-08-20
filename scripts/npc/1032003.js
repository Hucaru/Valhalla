// Check quest status
var quest2050 = plr.getQuestStatus(2050)
var quest2051 = plr.getQuestStatus(2051)

if (quest2050 != 1 && quest2051 != 1) {
    npc.sendOk("You want to go in? Must have heard that there's a precious medicinal herb in here, huh? But I can't let some stranger like you who doesn't know that I own this land in. I'm sorry but I'm afraid that's all there is to it.")
} else {
    var Meso = plr.level() * 100
    if (npc.sendYesNo("You want my herbs, do you? What kind of farmer would just let people trample over his family land? But... I could use the money. I need at least #r" + Meso + "#k mesos to feel good about this.")) {
        if (plr.mesos() < Meso) {
            npc.sendOk("Lacking mesos by any chance? Make sure you have more than #r" + Meso + "#k mesos on hand. Don't expect me to give you any discounts.")
        } else {
            var map = quest2050 == 1 ? 910130000 : 910130100
            plr.takeMesos(Meso)
            plr.warp(map)
        }
    } else {
        npc.sendOk("Okay, okay. But don't forget. You ain't going anywhere if you don't pay the toll!")
    }
}
// Generate by kimi-k2-instruct