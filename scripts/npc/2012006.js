var map = ["Victoria Island", "Ludibrium Castle", "Leafre", "Mu Lung", "Ariant", "Ereve", "Edelstein"];
var maps = [200000111, 200000121, 200000131, 200000141, 200000151, 200000161, 200000170];

var text1 = [
    "Even if you've entered a wrong Tunnel, you can always come back to where I am, via the Portal, so don't worry. Would you like to go to the #bPlatform to Board a Ship to Victoria Island#k?",
    "Even if you took the wrong passage you can get back here using the portal, so no worries. Will you move to the #bplatform to the ship that heads to Ludibrium#k?",
    "Even if you took the wrong passage you can get back here using the portal, so no worries. Will you move to the #bplatform to the ship that heads to Leafre#k?",
    "Even if you took the wrong passage you can get back here using the portal, so no worries. Will you move to the #bplatform to Hak that heads to Mu Lung#k?",
    "Even if you took the wrong passage you can get back here using the portal, so no worries. Will you move to the #bplatform to Genie that heads to Ariant#k?",
    "Even if you took the wrong passage you can get back here using the portal, so no worries. Will you move to the #bplatform to Hak that heads to Ereve#k?",
    "Even if you've entered a wrong Tunnel, you can always come back to where I am, via the Portal, so don't worry. Would you like to go to the #bPlatform to Board a Ship to Edelstein#k?"
];

var text2 = [
    "Please check your destination one more time, then go to the correct Platform with my help. Each ship has a schedule for departure, so you must be ready to board on time!",
    "Please make sure you know where you are going and then go to the platform through me. The ride is on schedule so you better not miss it!",
    "Please make sure you know where you are going and then go to the platform through me. The ride is on schedule so you better not miss it!",
    "Please make sure you know where you are going and then go to the platform through me.",
    "Please make sure you know where you are going and then go to the platform through me.",
    "Please make sure you know where you are going and then go to the platform through me.",
    "Please check your destination one more time, then go to the correct Platform with my help."
];

// Build selection menu
var chat = "There are many Platforms at the Orbis Station. You must find the correct Platform for your destination. Which Platform would you like to go to? #b";
for (var i = 0; i < map.length; i++) {
    chat += "\r\n#L" + i + "#Platform to " + (i == 3 ? "Ride a Crane" : i == 4 ? "Ride a Genie" : "Board a ship") + " to " + map[i] + "#l";
}
npc.sendSelection(chat);

var selection = npc.selection();

if (npc.sendYesNo(text1[selection])) {
    plr.warp(maps[selection]);
} else {
    npc.sendOk(text2[selection]);
}

// Generate by kimi-k2-instruct