// Shane – Ellinia gatekeeper

var q_dietMed = 2050;  // "Sabitrama and the Diet Medicine"
var q_agingMed = 2051; // "Sabitrama's Anti-Aging Medicine"

// Status: 0 = none, 1 = in-progress, 2 = completed
var dietStatus  = plr.getQuestStatus(q_dietMed);
var dietRecord  = (plr.quest(q_dietMed).data || "");
var agingStatus = plr.getQuestStatus(q_agingMed);
var agingRecord = (plr.quest(q_agingMed).data || "");
var lvl = plr.getLevel();

// Split state into non-overlapping buckets
var doingDiet          = (dietStatus === 1 && lvl >= 25);
var hasDietUnlockedOnly = (dietRecord === "1_00" && lvl >= 25); // record-based gate only

var doingAgingAdv        = (agingStatus === 1 && lvl >= 50);
var hasAgingUnlockedOnly = (agingRecord === "2_00" && lvl >= 50);

// Choose ONE path per interaction using an else-if ladder

// 1) Diet quest in progress (paid entry)
if (doingDiet) {
    if (npc.sendYesNo(
        "So you came here at the request of #b#p1061005##k to take the medicinal herb? Well... " +
        "I inherited this land from my father and I can't let some stranger in just like that... " +
        "But, with #r3400#k mesos, it's a whole different story. Pay and enter?"
    )) {
        if (plr.mesos() >= 3400) {
            plr.giveMesos(-3400);
            plr.warp(101000100);
        } else {
            npc.sendOk("Are you missing money? Make sure you have at least #r3400#k mesos on hand.");
        }
    } else {
        npc.sendOk("I understand... but I can’t let you in for free.");
    }

// 2) Anti-aging quest in progress (paid entry deeper in)
} else if (doingAgingAdv) {
    if (npc.sendYesNo(
        "It's you from the other day... need to go further? It's dangerous there, but for #r10000#k mesos, " +
        "I'll let you search through everything. Pay and enter?"
    )) {
        if (plr.mesos() >= 10000) {
            plr.giveMesos(-10000);
            plr.warp(101000102);
        } else {
            npc.sendOk("You're short on mesos. You need #r10000#k.");
        }
    } else {
        npc.sendOk("Alright, maybe next time.");
    }

// 3) Anti-aging unlocked by record (free entry deeper map)
} else if (hasAgingUnlockedOnly) {
    npc.sendBackNext(
        "It's you from the other day... is #b#p1061005##k working hard on the anti-aging medicine? " +
        "You’ve proven yourself, so I’ll let you pass without charge.",
        true, true
    );
    npc.sendBackNext(
        "By the way, #b#p1032100##k tried to sneak in earlier and dropped something inside. " +
        "I couldn’t find it. Maybe you’ll have better luck.",
        true, true
    );
    plr.warp(101000102);

// 4) Diet unlocked by record (free entry first map)
} else if (hasDietUnlockedOnly) {
    npc.sendBackNext(
        "It's you from the other day... is #b#p1061005##k working hard on the diet medicine? " +
        "I was impressed you made it this far; you can enter for free.",
        true, true
    );
    plr.warp(101000100);

// 5) Default greeting
} else {
    npc.sendOk(
        "Do you want to enter this place? There are rare herbs inside, but I can’t let strangers onto my property. " +
        "Sorry, but you’ll have to leave."
    );
}