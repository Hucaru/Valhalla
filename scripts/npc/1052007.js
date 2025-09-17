// Ticketing Gate NPC – stateless rewrite

var item  = [4031036, 4031037, 4031038];
var map   = [910360000, 910360100, 910360200];

function pickerLine(questId, testStatus) {
    return plr.getQuestStatus(questId) > 0 && plr.getQuestStatus(1602) < 2 ? "\r\n#L0##eSubway Construction Site#n#l" : "";
}

var choice = npc.sendMenu(
      "Pick your destination. #b" + pickerLine(1600, 0)
    + "\r\n#L1#Kerning city Subway#r(Beware of Stirges and Wraiths!)#l"
    + "\r\n#L2##bKerning Square Shopping Center (Get on the Subway).#l"
    + "\r\n\r\n#L3#Enter Construction Site#l"
    + "\r\n#L4#New Leaf City#l");

if (choice === 0) {                        // ---- q1600 construction ----
    var em = cm.getEventManager("q1600");
    var prop = em.getProperty("state");
    if (prop === null || prop === "0") {
        em.startInstance(plr);
    } else {
        cm.getClient().getSession().write(
            Packages.tools.packet.MaplePacketCreator.serverNotice(
                5,
                "Someone is already in the Subway Construction Site. Please try again later."));
    }
} else if (choice === 1) {                 // ---- Kerning city Subway ----
    plr.changeMap(cm.getMap(103020100), cm.getMap(103020100).getPortal(2));
} else if (choice === 2) {                 // ---- Kerning Square ----
    plr.changeMap(cm.getMap(103020010), cm.getMap(103020010).getPortal(0));
    cm.getClient().getSession().write(
        Packages.tools.packet.MaplePacketCreator.topMsg(
            "The next stop is at Kerning Square Station. The exit is to your left."));
    cm.getClient().getSession().write(
        Packages.tools.packet.MaplePacketCreator.serverNotice(
            6,
            "The next stop is at Kerning Square Station. The exit is to your left."));
    plr.startMapTimeLimitTask(10, cm.getMap(103020020));
} else if (choice === 3) {                 // ---- Construction site tickets ----
    var ok = false;
    for (var i = 0; i < item.length; i++) {
        if (plr.itemCount(item[i]) > 0) { ok = true; break; }
    }
    if (!ok) {
        npc.sendOk("Here's the ticket reader. You are not allowed in without the ticket.");
    } else {
        var chat = "Here's the ticket reader. You will be brought in immediately. Which ticket would you like to use? #b";
        for (var j = 0; j < item.length; j++)
            if (plr.itemCount(item[j]) > 0)
                chat += "\r\n#L" + j + "# Construction site B" + (j + 1) + "#l";
        var siteSel = npc.sendMenu(chat);
        if (plr.itemCount(item[siteSel]) > 0) {
            plr.removeItem(item[siteSel], 1);
            plr.changeMap(cm.getMap(map[siteSel]), cm.getMap(map[siteSel]).getPortal(0));
        }
    }
} else if (choice === 4) {                 // ---- New Leaf City ----
    if (plr.itemCount(4031711) < 1) {
        npc.sendOk("Here's the ticket reader. You are not allowed in without the ticket.");
    } else {
        var subwaySel = npc.sendMenu(
            "Here's the ticket reader. You will be brought in immediately. Which ticket would you like to use? \r\n#L1##bNew Leaf city (Normal)#l");
        if (subwaySel === 1) {
            var em2 = cm.getEventManager("Subway");
            if (em2.getProperty("entry") === "false" && em2.getProperty("docked") === "true") {
                npc.sendOk(
                    "We will begin boarding 1 minutes before the takeoff. Please be patient and wait for a few minutes. "
                    + "Be aware that the subway will take off right on time, and we stop receiving tickets 1 minute before that, "
                    + "so please make sure to be here on time.");
            } else if (em2.getProperty("entry") === "false") {
                npc.sendOk(
                  "The subway for New Leaf city is preparing for takeoff. "
                  + "I'm sorry, but you'll have to hop on the next ride. "
                  + "The ride schedule is available through the usher at the ticketing booth.");
            } else if (npc.sendYesNo(
                  "It looks like there's plenty of room for this ride. "
                  + "Please have your ticket ready so I can let you in. "
                  + "The ride will be long, but you'll get to your destination just fine. "
                  + "What do you think? Do you want to get on this ride?")) {
                plr.removeItem(4031711, 1);
                plr.changeMap(cm.getMap(600010004), cm.getMap(600010004).getPortal(0));
            }
        }
    }
}