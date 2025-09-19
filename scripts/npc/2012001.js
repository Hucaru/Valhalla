// Orbis -> Ellinia station boarding
var TICKET_ID = 4031047;
var BOARD_MAP = 200000112;

var props = plr.instanceProperties();
var salesOpen = ("canSellTickets" in props) && props["canSellTickets"];

if (!salesOpen) {
    npc.sendOk("Boarding is not available right now. Please wait for the next boarding announcement.");
} else if (plr.itemCount(TICKET_ID) < 1) {
    npc.sendOk(
        "You need an Ellinia Ticket to board.\r\n" +
        "Please purchase #t" + TICKET_ID + "# at the ticket counter and come back."
    );
} else {
    var go = npc.sendYesNo(
        "Boarding for the airship to Ellinia is now open.\r\n" +
        "Your ticket will be collected upon entry. Would you like to board now?"
    );
    if (go) {
        plr.removeItemsByID(TICKET_ID, 1);
        plr.warp(BOARD_MAP);
    } else {
        npc.sendOk("All right. Please let me know when you are ready to board.");
    }
}
