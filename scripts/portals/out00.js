(function () {
    const FM_ID = 910000000;
    const FALLBACK_ID = 102000000;

    var prev = (plr.previousMap && plr.previousMap()) || 0;
    if (prev && prev > 0 && prev !== FM_ID) {
        plr.warpFromName(prev, "st00");
        return;
    }

    plr.warpFromName(FALLBACK_ID, "st00");
})();
