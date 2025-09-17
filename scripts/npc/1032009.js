// Ticketing Usher 1032009 – Ellinia to Orbis
if (npc.sendYesNo("We're just about to take off. Are you sure you want to get off the ship? You may do so, but then you'll have to wait until the next available flight, also, the ticket is NOT refundable. Do you still wish to get off board?")) {
    plr.warp(101000300);
} else {
    npc.sendOk("You'll get to your destination in a short while. Talk to other passengers and share your stories to them, and you'll be there before you know it.");
}