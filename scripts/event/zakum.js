// Zakum
var BOSS_MAP = 280030000;
var LEAVE_MAP = 211042300;
var MAX_DURATION = 60 * 60 * 1000;
var TICK_MS    = 5   * 1000;

function init(controller) {
    controller.log("Zakum event script loaded");
    controller.schedule("tick", 0);
}

function tick(controller) {
    var field = controller.field(BOSS_MAP);
    if (!field) {
        controller.schedule("tick", TICK_MS);
        return;
    }
    var props = field.getProperties(0);
    if (!props) {
        props = field.getProperties(0);
        props.eventActive = false;
    }

    var players = field.playerCount(0);
    var mobs    = field.mobCount(0);

    if (!props.eventActive) {
        if (mobs > 0 && players === 0) {
            controller.log("Zakum event – quick reset");
            field.reset();
            controller.schedule("tick", TICK_MS);
            return;
        }

        if (mobs > 0 && players > 0) {
            props.eventActive   = true;
            props.eventStarted  = Date.now();
            props.hardTimerFired = false;
            controller.log("Zakum event started");
            controller.schedule("hardTimeout", MAX_DURATION);
        }
    } else {

        if (props.hardTimerFired) {
            field.warpPlayersToPortal(LEAVE_MAP, 0);
        }

        if (players === 0) {
            controller.log("Zakum event – resetting map");
            field.reset();
            props.eventActive = false;
        }
    }

    controller.schedule("tick", TICK_MS);
}

function hardTimeout(controller) {
    var props = controller.field(BOSS_MAP).getProperties(0);
    if (props && props.eventActive) {
        props.hardTimerFired = true;
        controller.log("Zakum event 60-min timeout – evicting players");
    }
}
