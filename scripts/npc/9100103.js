// Gashapon NPC – neck-breaking to stateless

var nextMapId = 100000006;     // hidden return default; restore warp() for clean close

// ---- 1. Check ticket presence ----
if (plr.itemCount(5220000) < 1 && plr.itemCount(5451000) < 1) {
    npc.sendOk("You don't have a single ticket with you. Please buy the ticket at the department store before coming back to me. Thank you.");
}

// ---- 2. Ask player if they want to spin ----
if (!npc.sendYesNo("You have some #bGachapon Tickets#k there. \r\nWould you like to try your luck?")) {
    npc.sendOk("Feel free to come back whenever you’re ready.");
}

// ---- 3. Consume one ticket ----
var ticketId = (plr.itemCount(5220000) > 0 && plr.getMapId() === 103000000)
        ? 5220000
        : 5451000;
plr.removeItemsByID(ticketId, 1);

// ---- 4. Randomly pick and attempt to give prize ----
var prizeList = [
    2040402, 2022130, 4130014, 2000004, 2000005,
    2022113, 1322008, 1302021, 1322022, 1302013,
    1051010, 1060079, 1002005, 1002023, 1002085,
    1332017, 1322010, 1051031, 1002212, 1002117,
    1040081, 1051037, 1472026, 1332015, 1041060,
    1472003, 1060086, 1060087, 1472009, 1060051,
    1041080, 1041106, 1092018
];
var pick = prizeList[Math.floor(Math.random() * prizeList.length)];

if (plr.giveItem(pick, 1)) {
    npc.sendOk("You have obtained #b#t" + pick + "##k.");
} else {
    // i.e., inventory full; however ticket already taken (cannot roll back simply)
    npc.sendOk("Please check your item inventory and see if you have the ticket, or if the inventory is full.");
}