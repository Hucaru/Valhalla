// Stage 8 NPC - LudiPQ (Box Combination Puzzle)
// Players must stand on 5 correct boxes out of 9 total boxes

var props = map.properties();

if (!plr.isPartyLeader()) {
    npc.sendOk("Here is information about the 8th stage. Here you will find many platforms to climb. #b5#k of them will be connected to the portal that leads to the next stage. To pass, place #b5 of your party members on the correct platform#k.\r\nA word of warning: You will need to stand firmly in the center of the platform for your answer to count as correct. Also remember that only 5 members can stay on the platform. When this happens, the party leader must #bclick me twice to know if the answer is correct or not#k. Good luck!");
} else {
    var totalAreas = 9; // 9 box positions
    var requiredPlayers = 5; // Need 5 players on correct boxes
    
    // Get or initialize the answer
    var answer = props.ans;
    if (!answer) {
        // Generate random answer: 5 correct positions (1) and 4 incorrect (0)
        var pattern = "111110000";
        var arr = pattern.split('');
        // Shuffle
        for (var i = arr.length - 1; i > 0; i--) {
            var j = Math.floor(Math.random() * (i + 1));
            var temp = arr[i];
            arr[i] = arr[j];
            arr[j] = temp;
        }
        answer = arr.join('');
        props.ans = answer;
    }
    
    // Check current player positions
    var currentPattern = "";
    var totalPlayers = 0;
    for (var i = 0; i < totalAreas; i++) {
        var count = map.playersInArea(i);
        if (count > 0) {
            currentPattern += "1";
            totalPlayers += count;
        } else {
            currentPattern += "0";
        }
    }
    
    // Validate the attempt
    if (totalPlayers < requiredPlayers) {
        npc.sendOk("Looks like you still haven't found the " + requiredPlayers + " correct platforms. You need to have " + requiredPlayers + " members of your party on the platforms, standing in the center. Currently only " + totalPlayers + " players are positioned. Keep trying!");
    } else if (totalPlayers > requiredPlayers) {
        npc.sendOk("Too many players on the platforms! You need exactly " + requiredPlayers + " members, but there are " + totalPlayers + " players positioned. Please adjust and try again!");
    } else if (currentPattern != answer) {
        map.showEffect("quest/party/wrong");
        map.playSound("Party1/Failed");
        npc.sendOk("That's not the correct combination! Try again. Remember, you need exactly " + requiredPlayers + " members standing on the correct platforms in the center.");
    } else {
        // Correct combination!
        props.clear = true;
        map.showEffect("quest/party/clear");
        map.playSound("Party1/Clear");
        map.portalEffect("gate");
        npc.sendOk("Perfect! You found the correct combination! The portal to the next stage is now open!");
    }
}
