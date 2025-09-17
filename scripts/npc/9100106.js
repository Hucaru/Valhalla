// Gachapon (Male Bath) – stateless rewrite

var prizes = [2000004, 2022113, 2040019, 2040020, 1072238, 1040081, 1382002, 1442021, 1072239, 1002096, 1322010, 1472005, 1002021, 1422007, 1082148, 1102081, 1040043, 1002117, 1302013, 1462024, 1382003, 1051001, 1472000, 1002088, 1472003, 1002048, 1002178, 1040007, 1002131, 1002288, 1002183, 1372006, 1442004, 1040082, 1322003, 2022195, 1412001, 1472009, 1060088, 1002035, 1322009, 1472016, 1332011, 1032027, 1002214, 1312014, 1002120, 1322023, 1452010, 1002034, 1060025, 1082147, 1002055, 1060019, 1002180, 1002154, 1060068, 1462013, 1022060, 1022058, 1012078, 1012079, 1012076];

if (plr.itemCount(5220000) < 1 && plr.itemCount(5451000) < 1) {
    npc.sendOk("You don't have a single ticket with you. Please buy the ticket at the department store before coming back to me. Thank you.");
} else {
    if (npc.sendYesNo("You have some #bGachapon Tickets#k there. \r\nWould you like to try your luck?")) {
        // Roll against in-memory prize list; pick one random
        var roll = prizes[Math.floor(Math.random() * prizes.length)];
        if (plr.giveItem(roll, 1)) {
            var ticket = plr.itemCount(5220000) > 0 ? 5220000 : 5451000;
            plr.removeItemsByID(ticket, 1);
            npc.sendOk("You have obtained #b#t" + roll + "##k.");
        } else {
            npc.sendOk("Please check your item inventory and see if you have the ticket, or if the inventory is full.");
        }
    }
}