// Zakum Party Quest (Stage 1 - Dead Mine)
// Maps for the quest
var maps = [280010000, 280010010, 280010020, 280010030];
var exitMapID = 211042300; // El Nath: Door to Zakum
var itemKeys = 4001016;     // Keys collected during PQ
var itemDocs = 4001015;     // Documents collected during PQ

function start() {
    ctrl.setDuration("30m");

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
    // Allow portals to be used freely in this PQ
    return true;
}

function afterPortal(plr, dst) {
    plr.showCountdown(ctrl.remainingTime());
}

function timeout(plr) {
    // Clean up items on timeout
    plr.removeItemsByID(itemKeys, plr.itemCount(itemKeys));
    plr.removeItemsByID(itemDocs, plr.itemCount(itemDocs));
    plr.warp(exitMapID);
}

function playerLeaveEvent(plr) {
    // Clean up items when player leaves
    plr.removeItemsByID(itemKeys, plr.itemCount(itemKeys));
    plr.removeItemsByID(itemDocs, plr.itemCount(itemDocs));

    ctrl.removePlayer(plr);
    plr.warp(exitMapID);

    // If party leader leaves or only one player left (or fewer), end the event
    if (plr.isPartyLeader() || ctrl.playerCount() <= 1) {
        var players = ctrl.players();

        for (let i = 0; i < players.length; i++) {
            players[i].removeItemsByID(itemKeys, players[i].itemCount(itemKeys));
            players[i].removeItemsByID(itemDocs, players[i].itemCount(itemDocs));
            players[i].warp(exitMapID);
        }

        ctrl.finished();
    }
}
