var maps = [922010100, 922010200, 922010201, 922010300, 922010400, 922010401, 922010402, 922010403, 922010404, 922010405, 922010500, 922010501, 922010502, 922010503, 922010504, 922010505, 922010506, 922010600, 922010700, 922010800, 922010900, 922011000, 922011100];
var exitMapID = 221024500; // Exit map
var pass = 4001022; // Pass of Dimension
var key = 4001023; // Key of Dimension

function start() {
    ctrl.setDuration("60m");

    for (let i = 0; i < maps.length; i++) {
        var field = ctrl.getMap(maps[i]);
        field.removeDrops();
        field.clearProperties();
    }

    var players = ctrl.players();
    var time = ctrl.remainingTime();

    for (let i = 0; i < players.length; i++) {
        players[i].warp(maps[0]);
        players[i].showCountdown(time);
    }
}

function beforePortal(plr, src, dst) {
    var props = src.properties();

    if (props["clear"]) {
        return true;
    }

    plr.sendMessage("Cannot use portal at the moment");
    return false;
}

function afterPortal(plr, dst) {
    plr.showCountdown(ctrl.remainingTime());

    var props = dst.properties();

    if (props["clear"]) {
        // send the active portal effect in case we have entered map after party cleared
        plr.portalEffect("gate");
    }
}

function timeout(plr) {
    plr.warp(exitMapID);
}

function playerLeaveEvent(plr) {
    plr.removeItemsByIDSilent(pass, plr.itemCount(pass));
    plr.removeItemsByIDSilent(key, plr.itemCount(key));

    ctrl.removePlayer(plr);
    plr.warp(exitMapID);

    if (plr.isPartyLeader() || ctrl.playerCount() < 3) {
        var players = ctrl.players();

        for (let i = 0; i < players.length; i++) {
            players[i].warp(exitMapID);
        }

        ctrl.finished();
    }
}
