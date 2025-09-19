// Ludi -> Orbis station boarding
var TICKET_ID = 4031045;
var BOARD_MAP = 220000111;

var props = plr.instanceProperties();
var boardingOpen = ("canBoard" in props) && props["canBoard"];

if (!boardingOpen) {
    npc.sendOk("Boarding is not available right now. Please wait for the next boarding announcement.");
} else if (plr.itemCount(TICKET_ID) < 1) {
    npc.sendOk(
        "You need an Orbis Ticket to board.\r\n" +
        "Please purchase #t" + TICKET_ID + "# at the ticket counter and come back."
    );
} else {
    var go = npc.sendYesNo(
        "Boarding for the airship to Orbis is now open.\r\n" +
        "Your ticket will be collected upon entry. Would you like to board now?"
    );
    if (go) {
        plr.removeItemsByID(TICKET_ID, 1);
        plr.warp(BOARD_MAP);
    } else {
        npc.sendOk("All right. Please let me know when you are ready to board.");
    }
}
