// Platform Service Manager - Orbis Station / 200000100
var mapNames = ["Victoria Island", "Ludibrium Castle", "Leafre", "Mu Lung", "Ariant", "Ereve", "Edelstein"];
var mapPortals = [200000111, 200000121, 200000131, 200000141, 200000151, 200000161, 200000170];

// Build platform-choice menu
var chat = "There are many Platforms at the Orbis Station. You must find the correct Platform for your destination. Which Platform would you like to go to? #b";
for (var i = 0; i < mapNames.length; i++) {
    chat += "\r\n#L" + i + "#Platform to ";
    if (i === 3) chat += "Ride a Crane";
    else if (i === 4) chat += "Ride a Genie";
    else chat += "Board a ship";
    chat += " to " + mapNames[i] + "#l";
}

// Prompt user for destination platform
npc.sendSelection(chat);
var sel = npc.selection();

// Compose confirmation prompt
var confirmText = "";
switch (sel) {
    case 0: confirmText = "Even if you've entered a wrong Tunnel, you can always come back to where I am, via the Portal, so don't worry. Would you like to go to the #bPlatform to Board a Ship to Victoria Island#k?"; break;
    case 1: confirmText = "Even if you took the wrong passage you can get back here using the portal, so no worries. Will you move to the #bplatform to the ship that heads to Ludibrium#k?"; break;
    case 2: confirmText = "Even if you took the wrong passage you can get back here using the portal, so no worries. Will you move to the #bplatform to the ship that heads to Leafre#k?"; break;
    case 3: confirmText = "Even if you took the wrong passage you can get back here using the portal, so no worries. Will you move to the #bplatform to Hak that heads to Mu Lung#k?"; break;
    case 4: confirmText = "Even if you took the wrong passage you can get back here using the portal, so no worries. Will you move to the #bplatform to Genie that heads to Ariant#k?"; break;
    case 5: confirmText = "Even if you took the wrong passage you can get back here using the portal, so no worries. Will you move to the #bplatform to Hak that heads to Ereve#k?"; break;
    case 6: confirmText = "Even if you've entered a wrong Tunnel, you can always come back to where I am, via the Portal, so don't worry. Would you like to go to the #bPlatform to Board a Ship to Edelstein#k?"; break;
    default: confirmText = "Please check your destination one more time, then go to the correct Platform with my help. Each ship has a schedule for departure, so you must be ready to board on time!";
}

// Ask confirmation; clicked “Yes” → warp, “No” → display reminder
if (npc.sendYesNo(confirmText)) {
    plr.warp(mapPortals[sel]);
} else {
    var remind = "";
    switch (sel) {
        case 0: remind = "Please check your destination one more time, then go to the correct Platform with my help. Each ship has a schedule for departure, so you must be ready to board on time!"; break;
        case 1: remind = "Please make sure you know where you are going and then go to the platform through me. The ride is on schedule so you better not miss it!"; break;
        case 2: remind = "Please make sure you know where you are going and then go to the platform through me. The ride is on schedule so you better not miss it!"; break;
        case 3: remind = "Please make sure you know where you are going and then go to the platform through me."; break;
        case 4: remind = "Please make sure you know where you are going and then go to the platform through me."; break;
        case 5: remind = "Please make sure you know where you are going and then go to the platform through me."; break;
        case 6: remind = "Please check your destination one more time, then go to the correct Platform with my help."; break;
    }
    npc.sendOk(remind);
}