var quest = [1431, 1432, 1433, 1435, 1436, 1437, 1439, 1440, 1442, 1443, 1445, 1446, 1447, 1448];
var y = true;

for (var i = 0; i < quest.length; i++)
    if (plr.getQuestStatus(quest[i]) === 1) {
        y = false;
    }

if (y) {
    npc.sendOk("#b(A mysterious energy surrounds this stone. It has an incredibly cold aura...)");
} else {
    if (npc.sendYesNo("#b(A mysterious energy surrounds this stone. The elder definetly told me to touch it... Should I really touch this thing?)")) {
        plr.warp(910540000);
    }
}