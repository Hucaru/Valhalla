// Ellinia station ticket seller
var TICKET_ID = 4031045; // Orbis Ticket
var PRICE = 5000;        // Mesos

npc.sendBackNext(
    "Hello there. Planning a trip to Orbis?\r\n" +
    "If you need a boarding ticket, I can help you with that. The airship staff will not let you on without one.",
    false, true
);

var menu =
    "What would you like to do?\r\n" +
    "#L0##bPurchase an Orbis Ticket (" + PRICE + " mesos)#k#l\r\n" +
    "#L1#What is this ticket for?#l\r\n" +
    "#L2#Never mind.#l";
npc.sendSelection(menu);

var sel = npc.selection();

if (sel == 0) {
    var yes = npc.sendYesNo(
        "An Orbis Ticket costs " + PRICE + " mesos.\r\n" +
        "Once you have it, head to the boarding platform and show your ticket to the usher.\r\n\r\n" +
        "Would you like to purchase one now?"
    );

    if (yes) {
        if (plr.mesos() < PRICE) {
            npc.sendOk("It looks like you do not have enough mesos. You will need " + PRICE + " mesos to buy the ticket.");
        } else {
            npc.sendOk(
                "Here is your ticket: #t" + TICKET_ID + "#.\r\n" +
                "Keep it safe and show it at the boarding platform to travel to Orbis. Safe travels."
            );

            if (plr.giveItem(TICKET_ID, 1)) {
                plr.giveMesos(-PRICE);
            } else {
            }
        }
    } else {
        npc.sendOk("No problem. Take your time and let me know if you decide to travel.");
    }

} else if (sel == 1) {
    npc.sendBackNext(
        "The Orbis Ticket grants you access to the airship that departs from the station platform.\r\n" +
        "Present it to the usher at the gate when boarding starts. Please arrive a little early.",
        true, true
    );
    npc.sendOk("If you decide to go, come back and I will get you a ticket.");

} else {
    npc.sendOk("If you change your mind, I will be here to help with your travel plans.");
}