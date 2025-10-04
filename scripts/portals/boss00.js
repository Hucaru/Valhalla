// Pianus Portal
// 230040410 -> 230040420

(function () {
    const destMap = 230040420;

    if (plr.playerCount(destMap) < 10) {
        plr.warpFromName(destMap, "out00");
    }
})();

