var maps = [103000800, 103000801, 103000802, 103000803, 103000804, 103000805];
var exitMapID = 103000890;
var coupon = 4001007;
var pass = 4001008;

function start() {
    ctrl.setDuration("30m");

    for (let i = 0; i < maps.length; i++) {
        var field = ctrl.getMap(maps[i]);
        field.removeDrops();
        field.clearProperties();
        // props = field.properties();
        // props["clear"] = false;
        // props["wrong"] = false;
    }

    var players = ctrl.players();
    var time = ctrl.remainingTime();

    for (let i = 0; i < players.length; i++) {
        players[i].warp(maps[0]);
        players[i].showCountdown(time);
    }
}

function beforePortal(plr, src, dst) {
    props = src.properties();

    if (props["clear"]) {
        return true;
    }

    plr.sendMessage("Cannot use portal at the moment");
    return false;
}

function afterPortal(plr, dst) {
    plr.showCountdown(ctrl.remainingTime());

    props = dst.properties();

    if (props["clear"]) {
        // send the active portal effect in case we have entered map after party cleared
        plr.portalEffect("gate");
    }
}

function timeout(plr) {
    plr.warp(exitMapID);
}

function playerLeaveEvent(plr) {
    plr.removeItemsByID(coupon, plr.itemCount(coupon));
    plr.removeItemsByID(pass, plr.itemCount(pass));

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