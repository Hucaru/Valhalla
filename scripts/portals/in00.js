// Pap Portal
// 220080000 -> 220080001
(function () {
    const BOSS_MAP = 220080001;

    if (plr.eventActive(BOSS_MAP)) {
        plr.sendMessage("The fight is already active. Please try again later.");
        return;
    }

    plr.warpFromName(BOSS_MAP, "st00");
})();
