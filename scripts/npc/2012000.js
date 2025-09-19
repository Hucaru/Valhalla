// Orbis station ticket seller
var TICKETS = [
    { id: 4031047, dest: "Ellinia",   price: 5000 },
    { id: 4031074, dest: "Ludibrium", price: 5000 }
];

npc.sendBackNext(
    "Welcome! Looking to travel by airship?\r\n" +
    "I can sell you boarding tickets for Ellinia or Ludibrium right here.",
    false, true
);

var menu = "What would you like to do?\r\n";
for (var i = 0; i < TICKETS.length; i++) {
    var tk = TICKETS[i];
    menu += "#L" + i + "##bPurchase #t" + tk.id + "# (" + tk.price + " mesos)#k#l\r\n";
}
menu += "#L9#What are these tickets?#l\r\n#L8#Never mind.#l";

npc.sendSelection(menu);
var sel = npc.selection();

if (sel >= 0 && sel < TICKETS.length) {
    var choice = TICKETS[sel];

    if (npc.sendYesNo(
        "You chose the ticket to " + choice.dest + ".\r\n" +
        "The price is " + choice.price + " mesos.\r\n\r\n" +
        "Would you like to purchase #t" + choice.id + "# now?"
    )) {
        if (plr.mesos() < choice.price) {
            npc.sendOk("You do not have enough mesos. You will need " + choice.price + " mesos to buy this ticket.");
        } else {
            npc.sendOk(
                "Here is your #t" + choice.id + "#.\r\n" +
                "Present it at the boarding gate when the airship is ready. Safe travels to " + choice.dest + "!"
            );

            if (plr.giveItem(choice.id, 1)) {
                plr.giveMesos(-choice.price);
            }
        }
    } else {
        npc.sendOk("All right. Take your time and let me know if you decide to travel.");
    }

} else if (sel === 9) {
    npc.sendBackNext(
        "These tickets allow you to board the airship bound for their respective destinations.\r\n" +
        "#t" + TICKETS[0].id + "# -> " + TICKETS[0].dest + "\r\n" +
        "#t" + TICKETS[1].id + "# -> " + TICKETS[1].dest + "\r\n\r\n" +
        "Keep the ticket in your inventory and show it to the usher at the platform when boarding starts.",
        true, true
    );
    npc.sendOk("If you decide to go, come back and I will sell you a ticket.");

} else {
    npc.sendOk("No problem. If you change your mind, I will be right here to help with your travel plans.");
}