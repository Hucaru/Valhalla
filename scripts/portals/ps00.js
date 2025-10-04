// Zakum Entrance Portal
// 211042300 -> 280030000
// There doesn't appear to be the sign-up room in v28... So we just send straight to Zakum's Altar
(function () {
    const BOSS_MAP = 280030000;

    if (plr.eventActive(BOSS_MAP)) {
        plr.sendMessage("The fight is already active. Please try again later.");
        return;
    }

    if (plr.playerCount(BOSS_MAP) > 20) {
        plr.sendMessage("The boss room is currently full.");
        return;
    }

    plr.warpFromName(BOSS_MAP, "west00");
})();
