// Zakum Party Quest (Stage 1 - Dead Mine)
// Maps for the quest
var maps = [280010000, 280010010, 280010011, 280010020, 280010030, 280010031, 280010040, 280010041, 280010050, 280010060, 280010070, 280010071, 280010080, 280010081, 280010090, 280010091, 280010100, 280010101, 280010110, 280010120, 280010130, 280010140, 280010150, 280011000, 280011001, 280011002, 280011003, 280011004, 280011005, 280011006]
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
    plr.removeItemsByIDSilent(itemKeys, plr.itemCount(itemKeys));
    plr.removeItemsByIDSilent(itemDocs, plr.itemCount(itemDocs));
    plr.warp(exitMapID);
}

function playerLeaveEvent(plr) {
    // Clean up items when player leaves
    plr.removeItemsByIDSilent(itemKeys, plr.itemCount(itemKeys));
    plr.removeItemsByIDSilent(itemDocs, plr.itemCount(itemDocs));

    ctrl.removePlayer(plr);
    plr.warp(exitMapID);

    // If party leader leaves or no players remain after removal, end the event
    if (plr.isPartyLeader() || ctrl.playerCount() <= 0) {
        var players = ctrl.players();

        for (let i = 0; i < players.length; i++) {
            players[i].removeItemsByIDSilent(itemKeys, players[i].itemCount(itemKeys));
            players[i].removeItemsByIDSilent(itemDocs, players[i].itemCount(itemDocs));
            players[i].warp(exitMapID);
        }

        ctrl.finished();
    }
}
