/*  ship status → stateless
    碼頭<前往維多利亞>    200000111
*/

var sk0 = plr.getSkillLevel(80001027);
var sk1 = plr.getSkillLevel(80001028);

if (sk0 != 1 && sk1 != 1) {
    // action0 path
    var goShip = npc.sendYesNo("Would you like to board the ship for #bVictoria Island#k now? It'll take about 30 seconds to get there.");
    if (goShip) {
        plr.warp(200090000);
        // server-side timer/map-limit handled elsewhere
    } else {
        npc.sendNext("Do you have some business you need to take care of here?");
    }
} else {
    // action1 path
    var sel = npc.sendMenu(
        "If you have an airplane, you can fly to stations all over the world. Would you rather take an airplane than wait for a ship? It'll cost you 5000 mesos.\r\n\r\n",
        "#b#L0#Use the airplane. #r(5000 mesos)#l",
        "#b#L1#Board a ship.#l"
    );

    if (sel === 1) {          // board ship
        var shipYes = npc.sendYesNo("Would you like to board the ship for #bVictoria Island#k now? It'll take about 30 seconds to get there.");
        if (shipYes) {
            plr.warp(200090000);
        } else {
            npc.sendNext("Do you have some business you need to take care of here?");
        }
    } else {                  // use airplane
        var plane = npc.sendMenu(
            "Which airplane would you like to use? #b",
            sk0 == 1 ? "\r\n#L0#Wooden Airplane#l" : "",
            sk1 == 1 ? "\r\n#L1#Rad Airplane#l" : ""
        );

        if (plr.mesos() >= 5000) {
            plr.giveMesos(-5000);
            plr.giveBuff(plane == 0 ? 80001027 : 80001028, 1);
            plr.warp(200110001);
        } else {
            npc.sendOk("You don't have enough money for the Station fee.");
        }
    }
}