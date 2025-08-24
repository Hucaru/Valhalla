// Demo NPC script (single-step options). 

// Config (use IDs that exist on your server)
var testQuestIdA = 3000; // If this doesn't exist in your data, "startQuest" will fail; script will force-add via setQuestData fallback.
var sampleMapId  = 100000000; // Henesys
var redPot       = 2000000;
var bluePot      = 2000001;
var sword        = 1302000;

function infoLine() {
    var qA = plr.quest(testQuestIdA);
    return ""
        + "Level: " + plr.getLevel() + "\r\n"
        + "Mesos: " + plr.mesos() + "\r\n"
        + "Job: " + plr.job() + "\r\n"
        + "QuestA: status=" + qA.status + " data=\"" + (qA.data || "") + "\"\r\n";
}

// Present a single selection menu; each option performs one action only
npc.sendSelection(
    "Function Demo (single-step)\r\n\r\n"
    + infoLine() + "\r\n"
    + "#L1#Add 1,000 mesos#l\r\n"
    + "#L2#Subtract 1,000 mesos#l\r\n"
    + "#L3#Give Red Pot x5#l\r\n"
    + "#L4#Give Sword x1#l\r\n"
    + "#L5#Warp to " + sampleMapId + "#l\r\n"
    + "#L6#Start Quest A (validate; fallback to force-add)#l\r\n"
    + "#L7#Set Quest A record to '1_00'#l\r\n"
    + "#L8#Complete Quest A#l\r\n"
    + "#L9#Forfeit Quest A#l\r\n"
);
var sel = npc.selection();

if (sel === 1) {
    plr.giveMesos(1000); // server updates mesos + sends stat packet
    npc.sendOk("Gave +1,000 mesos.\r\nNow: " + plr.mesos());
} else if (sel === 2) {
    plr.giveMesos(-1000); // subtract
    npc.sendOk("Took -1,000 mesos.\r\nNow: " + plr.mesos());
} else if (sel === 3) {
    var ok = plr.giveItem(redPot, 5); // server creates item, saves, sends inventory packet
    npc.sendOk(ok ? "Gave Red Pot x5." : "Failed to give item (invalid ID or no space).");
} else if (sel === 4) {
    var ok2 = plr.giveItem(sword, 1);
    npc.sendOk(ok2 ? "Gave Sword x1." : "Failed to give item (invalid ID or no space).");
} else if (sel === 5) {
    if (npc.sendYesNo("Warp to " + sampleMapId + " now?")) {
        plr.warp(sampleMapId);
    } else {
        npc.sendOk("Staying here.");
    }
} else if (sel === 6) {
    // Try normal start (validates against quest data). If it fails, force-add as in-progress using setQuestData.
    var okStart = plr.startQuest(testQuestIdA);
    if (okStart) {
        npc.sendOk("Quest A started (validated).");
    } else {
        if (plr.getQuestStatus(testQuestIdA) === 0) {
            plr.setQuestData(testQuestIdA, "");
            npc.sendOk("Quest A started (forced via record upsert).");
        } else {
            npc.sendOk("Quest A could not be started (requirements not met).");
        }
    }
} else if (sel === 7) {
    plr.setQuestData(testQuestIdA, "1_00"); // server upserts record + sends quest update packet
    npc.sendOk("Quest A record set to '1_00'.");
} else if (sel === 8) {
    var okC = plr.completeQuest(testQuestIdA); // validates reqs, moves to completed, sends completion packets
    npc.sendOk(okC ? "Quest A completed." : "Unable to complete Quest A (requirements not met).");
} else if (sel === 9) {
    if (plr.getQuestStatus(testQuestIdA) === 1) {
        plr.forfeitQuest(testQuestIdA); // removes in-progress, clears mob kills, sends remove packet
        npc.sendOk("Quest A forfeited.");
    } else {
        npc.sendOk("Quest A is not in-progress.");
    }
} else {
    npc.sendOk("No action.");
}