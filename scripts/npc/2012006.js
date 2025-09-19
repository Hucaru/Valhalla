// Platform Service Manager - Orbis Station / 200000100
var mapNames = ["Victoria Island", "Ludibrium"];
var mapPortals = [200000111, 200000121];

var chat = "There are two platforms at Orbis Station. Which platform would you like to go to? #b";
for (var i = 0; i < mapNames.length; i++) {
    chat += "\r\n#L" + i + "#Platform to board a ship to " + mapNames[i] + "#l";
}

npc.sendSelection(chat);
var sel = npc.selection();

if (sel < 0 || sel >= mapNames.length) {
    npc.sendOk("Please choose a valid platform. Each ship follows a schedule, so be ready to board on time.");
} else {
    var confirmText = "";
    if (sel === 0) {
        confirmText = "Even if you enter the wrong passage, you can return here using the portal. Move to the platform to board a ship to Victoria Island?";
    } else {
        confirmText = "Even if you enter the wrong passage, you can return here using the portal. Move to the platform to board a ship to Ludibrium?";
    }

    if (npc.sendYesNo(confirmText)) {
        plr.warp(mapPortals[sel]);
    } else {
        npc.sendOk("Please check your destination and use me to move to the correct platform. Do not miss the ship.");
    }
}
