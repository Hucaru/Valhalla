// Remove strict ordering and cm.dispose(), use eim straight
var puzzle = ["30*10+98", "3*8+610", "69+420", "400+140-72", "50*10+80-4", "5*60+5*5", "900/2+3", "20*20+15", "20*30+15", "9*9+100-43"];
var answer = ["001000011", "001101000", "000100011", "000101010", "000011100", "011010000", "001110000", "100110000", "100011000", "101000010"];

var eim = plr.getEventInstance();
if (!eim) {
    npc.sendOk("You are not in an event instance.");
}

// Determine which route
var stageStatus = eim.getProperty("stage5");
var isLeader = (plr.partyLeaderId() == plr.id());
var problemActive = (eim.getProperty("stage5a") !== null);

// Branching logic, stateless
if (!isLeader) {
    // follower path
    npc.sendNext("In the fifth stage, you will find a number of platforms. Of these platforms, #b3 are connected to the portal that leads to the next stage. 3 members of your party must stand in the center of these 3 platforms#k. \r\nRemember, exactly 3 members must be on a platform. No more, no less. While they are on the platform, the party leader must #bdouble-click on me to check whether the members have chosen the right platform#k. Good luck!");
}

// Leader flow
if (stageStatus === null || !problemActive) {
    // First time, create problem
    npc.sendBackNext("In the fifth stage, you will find a number of platforms. Of these platforms, #b3 are connected to the portal that leads to the next stage. 3 members of your party must stand in the center of these 3 platforms#k. \r\nRemember, exactly 3 members must be on a platform. No more, no less. While they are on the platform, the party leader must #bdouble-click on me to check whether the members have chosen the right platform#k. Good luck!", false, true);
    npc.sendBackNext("The #rthree numbers in the answer to my question are the key to opening the portal to the next stage. \r\n#r" + puzzle[num] + " = ?#k \r\nPlease find the correct answer.", true, true);    

    // set vars (only runs on first click)
    eim.setProperty("stage5", "0");
    var num = Math.floor(Math.random() * 10);
    eim.setProperty("stage5a", num.toString());
    plr.getMap().startMapEffect("" + puzzle[num] + " = ?", 5120018);

}

// Platform check if stage5 == 0
if (stageStatus === "0") {
    var count = 0;
    var x = "";
    for (var i = 0; i < 9; i++) {
        var n = plr.getMap().getNumPlayersItemsInArea(i);
        if (n > 0) count++;
        x += n;
    }
    if (count !== 3) {
        npc.sendNext("You haven't found the 3 correct platforms yet. Do you remember the question? I'll tell you again. \r\n#r" + puzzle[parseInt(eim.getProperty("stage5a"))] + " = ?#k \r\nRemember that only 3 party members should be on a platform, and they need to be standing in the center of the platform in order to be considered correct. Good luck!");

    }

    if (x === answer[parseInt(eim.getProperty("stage5a"))]) {
        eim.setProperty("stage5", "1");
        plr.getMap().broadcastEnvironmentChange("gate", 2);
        plr.getMap().broadcastEnvironmentChange("quest/party/clear", 3);
        plr.getMap().broadcastEnvironmentChange("Party1/Clear", 4);
    } else {
        plr.getMap().broadcastEnvironmentChange("quest/party/wrong_kor", 3);
        plr.getMap().broadcastEnvironmentChange("Party1/Failed", 4);
    }

}

if (stageStatus === "1") {
    npc.sendNext("Congratulations on clearing the quests for this stage. Please use the portal you see over there and move on to the next stage.");
}