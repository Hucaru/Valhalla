// Sauna Hotel (Sleepywood) – stateless, nested flow

npc.sendBackNext(
    "Welcome. We're the #m105000000# Hotel. Our hotel works hard to serve you the best at all times. If you are tired and worn out from hunting, how about a relaxing stay at our hotel?",
    false,
    true
);

var chat = "We offer two kinds of rooms for service. Please choose the one of your liking.";
chat += "\r\n#L0##b#m105000011# (499 mesos per use)#k#l\r\n";
chat += "#L1##b#m105000012# (999 mesos per use)#k#l";
npc.sendSelection(chat);

var select = npc.selection();

if (select !== 0 && select !== 1) {
    npc.sendOk("Please choose one of our rooms if you'd like to stay.");
} else {
    var isRegular = (select === 0);
    var cost = isRegular ? 499 : 999;

    var msg = isRegular
        ? "You've chosen the regular sauna. Your HP and MP will recover fast and you can even purchase some items there. Are you sure you want to go in?"
        : "You've chosen the VIP sauna. Your HP and MP will recover even faster than that of the regular sauna and you can even find a special item in there. Are you sure you want to go in?";

    if (npc.sendYesNo(msg)) {
        if (plr.mesos() < cost) {
            npc.sendOk("I'm sorry. It looks like you don't have enough mesos. It will cost you " + cost + " mesos to stay at our hotel.");
        } else {
            plr.takeMesos(cost);
            plr.warp(isRegular ? 105000011 : 105000012);
        }
    } else {
        npc.sendOk("We offer other kinds of services, too, so please think carefully and then make your decision.");
    }
}