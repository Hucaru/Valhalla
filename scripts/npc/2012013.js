// --------- stateless rewrite ----------
const skillWooden = 80001027;
const skillRad      = 80001028;

const hasWooden = plr.getSkillLevel(skillWooden) === 1;
const hasRad    = plr.getSkillLevel(skillRad) === 1;

let sel;

if (!hasWooden && !hasRad) {               // ---------- path 0 ----------
    npc.sendSelection(
        "Would you like to board the ship to Ludibrium? It take about 1 minute to arrive."
        + "\r\n#L0##bI'd like to board the ship.#l"
    );
    sel = npc.selection();

    if (npc.sendYesNo("Would you like to board the ship to Ludibrium now?")) {
        plr.warp(200090100, 0);
        plr.startMapTimeLimitTask(60, 220000110);
    } else {
        npc.sendOk("Do you have some business you need to take care of here?");
    }

} else {                                   // ---------- path 1 ----------
    npc.sendSelection(
        "If you have an airplane, you can fly to stations all over the world. Would you rather take an airplane than wait for a ship? It'll cost you 5,000 mesos to use the station."
        + "\r\n\r\n#b#L0#I'd like to use the plane. #r(5000 mesos)#l"
        + "\r\n#L1##bI'd like to board the ship.#l"
    );
    sel = npc.selection();

    if (sel === 0) {          // use the plane
        let planeSel;
        const builder = "Which airplane would you like to use? #b";
        let opts = "";
        if (hasWooden) opts += "\r\n#L0#Wooden Airplane#l";
        if (hasRad)    opts += "\r\n#L1#Rad Airplane#l";

        npc.sendSelection(builder + opts);
        planeSel = npc.selection();

        if (plr.mesos() > 5000) {
            plr.takeMesos(5000);
            plr.giveBuff(planeSel === 0 ? skillWooden : skillRad, 1);
            plr.warp(200110021, 0);
        } else {
            npc.sendOk("Please check and see if you have enough mesos to go.");
        }

    } else {                  // ride the ship
        if (npc.sendYesNo("Would you like to board the ship to Ludibrium now?")) {
            plr.warp(200090100, 0);
            plr.startMapTimeLimitTask(60, 220000110);
        } else {
            npc.sendOk("Do you have some business you need to take care of here?");
        }
    }
}