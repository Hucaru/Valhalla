// Fix prize list as in the original
var prize = [2044102, 2044502, 2041020, 2041017, 2041011, 2041014, 2044702, 2044602, 2043302, 2040302, 2040805, 2040502, 2044402, 2040902, 2040708, 2040402, 2043002, 2044101, 2041022, 2044701, 2040804, 2040702, 2040707, 2043801, 2044001, 2043701, 2048003, 2048000, 4020001, 4020002, 2060001, 2020002, 2012003, 4004002, 4020007, 2000004, 2012001, 2050003, 2020005, 4010006, 2020004, 2002002, 2020012, 2020009, 4010005, 2020003, 4004000, 2000005, 2020013, 2030000, 2030001, 2030002, 2030003, 2030004, 2030005, 2030006, 2030007, 2030019, 2020000, 2012002, 4020005, 4010004, 2020014, 4006001, 4006000, 2050002, 2002003, 1032003, 1302022, 1432009, 1102014, 1102018, 1322023, 1322025, 1032008, 1432008, 1322022, 1442018, 1442039, 1322027, 1032004, 1032026, 1442015, 1032016, 1032018, 1422004, 1422006, 1302021, 1322024, 1322012, 1051017, 1432015, 1032001, 1432018, 1432000, 1402014, 1032000, 1422000, 1032009, 1082145, 1452004, 1452000, 1002162, 1452003, 1040068, 1060057, 1002163, 1452001, 1002161, 1002038, 1002036, 1002013, 1372002, 1372003, 1040044, 1041048, 1041049, 1002150, 1472004, 1002175, 1472003, 1472002, 1041050, 1041047, 1332013, 1060019, 1442009, 1442006, 1422002, 1402003, 1040021, 1442010, 1002009, 1442007, 1442031, 1332002, 1060017, 2100000, 2044902, 2044901, 2044803, 2044804];

// Check for at least one ticket
if (plr.itemCount(5220000) < 1 && plr.itemCount(5451000) < 1) {
    npc.sendOk("You don't have a single ticket with you. Please buy the ticket at the department store before coming back to me. Thank you.");
    // script terminates naturally
} else {
    // Ask if they want to try
    if (npc.sendYesNo("You have some #bGachapon Tickets#k there. \r\nWould you like to try your luck?")) {
        // Determine which ticket to consume
        var ticket = (plr.itemCount(5220000) && plr.getMapId() === 105000000) ? 5220000 : 5451000;
        
        // Attempt to hand out a random prize
        var reward = prize[Math.floor(Math.random() * prize.length)];
        
        if (plr.giveItem(reward, 1)) {          // inventory has space → gave item OK
            plr.takeItem(ticket, -1, 1, -1);    // consume the ticket (type inv flags may vary by engine; adjust if needed)
            npc.sendOk("You have obtained #b#t" + reward + "##k.");
        } else {
            npc.sendOk("Please check your item inventory and see if you have the ticket, or if the inventory is full.");
        }
    }
}