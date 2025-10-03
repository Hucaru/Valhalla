// Papulatus
var BOSS_MAP = 220080001;
var LEAVE_MAP = 220080000;
var MAX_DURATION = 60 * 60 * 1000;
var TICK_MS    = 5   * 1000;

function init(controller) {
    controller.log("Papulatus event script loaded");
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
            controller.log("Papulatus event – quick reset");
            field.reset();
            controller.schedule("tick", TICK_MS);
            return;
        }

        if (mobs > 0 && players > 0) {
            props.eventActive   = true;
            props.eventStarted  = Date.now();
            props.hardTimerFired = false;
            controller.log("Papulatus event started");
            controller.schedule("hardTimeout", MAX_DURATION);
        }
    } else {
        if (props.hardTimerFired) {
            field.warpPlayersToPortal(LEAVE_MAP, 0);
        }

        if (players === 0) {
            controller.log("Papulatus event – resetting map");
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
        controller.log("Papulatus event 60-min timeout – evicting players");
    }
}
