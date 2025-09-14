var mapId = plr.getMapId();
var stagePart = ("" + mapId).substr(3,1);

if (plr.isPartyLeader()) {

    // ---------- Party leader side ----------

    if (stagePart === "1") {
        if (plr.instanceProperty("stage1b") === null) plr.instanceProperty("stage1b", 0);

        if (plr.instanceProperty("stage1a") === null) {
            npc.sendNext("Hello, and welcome the the first stage. As you can see, this place is full of Ligators. Each Ligator will drop one #bcoupon#k when defeated. Each party member, except the party leader, must come talk to me and then bring me the exact number of #bcoupons#k that I ask for. Once everyone #bcompletes their individual missions#k, the party can move on to the next stage. You must hurry, since the number of stages available depends on how fast you complete this stage. Good luck!");
            plr.instanceProperty("stage1a", 1);
        } else if (plr.instanceProperty("stage1a") === "1") {
            npc.sendNext("I'm sorry, but at least one party member still hasn't completed their mission. Everyone except the part leader must clear their mission to move on.");
        } else if (plr.instanceProperty("stage1a") === "2") {
            npc.sendNext("Congratulations on clearing this stage! I will create a portal that will lead you to the next one. You're on a time limit, so please hurry! Good luck!");
        } else {
            npc.sendNext("You all have cleared the quest for this stage. Use the portal to move to the next stage...");
        }
    }

    else if (stagePart === "2") {
        if (plr.instanceProperty("stage2a") === null) {
            npc.sendNext("Hi. Welcome to the 2nd stage. Next to me, you'll see a number of ropes. Out of these ropes, #b3 are connected to the portal that sends you to the next stage#k. All you need to do is have #b3 party members to find the answer ropes and hang on them#k. \r\nBUT, it doesn't count as an answer if you hang on to the rope too low; please bring yourself up enough to be counted as a correct answer. Also, only 3 members of your party are allowed on the ropes. Once they are hanging on, the leader of the party must #bdouble-click me to check and see if the answer's correct or not#k. Now, find the right ropes to hang on!");
            plr.instanceProperty("stage2a", 1);
        } else {
            if (plr.instanceProperty("stage2b") === null) {
                var rand = Math.random();
                var combination = (rand < 0.2) ? "1110" : (rand < 0.4) ? "1101" : (rand < 0.7) ? "1011" : "0111";
                plr.instanceProperty("stage2b", combination);
            }
            var correct = plr.instanceProperty("stage2b");
            var onRopes = 0;
            var pattern = "";
            for (var i = 0; i < 4; i++) {
                var count = plr.mapPlayersOnArea(i);
                pattern += count > 0 ? "1" : "0";
                onRopes += count > 0 ? 1 : 0;
            }
            if (onRopes !== 3) {
                npc.sendNext("It looks like you haven't found the 3 ropes just yet. Please think of a different combination of ropes, Only 3 are allowed to hang on to ropes, and if you hang on to too low, it won't count as an answer; so please keep that in mind. Keep going!");
            } else if (pattern === correct) {
                plr.instanceProperty("stage2", 1);
                npc.sendNext("Congratulations on clearing this stage! I will create a portal that will lead you to the next one. You're on a time limit, so please hurry! Good luck!");
            } else {
                plr.mapEffect("quest/party/wrong_kor");
                npc.sendNext("That is not the correct combination. Keep trying!");
            }
        }
    }

    else if (stagePart === "3") {
        if (plr.instanceProperty("stage3a") !== null) {
            if (plr.instanceProperty("stage3b") === null) {
                var rand = Math.random();
                var combination = (rand < 0.2) ? "11001" : (rand < 0.4) ? "01110" : (rand < 0.6) ? "10101" : (rand < 0.8) ? "10110" : "10011";
                plr.instanceProperty("stage3b", combination);
            }
            var correct = plr.instanceProperty("stage3b");
            var onPlatforms = 0;
            var pattern = "";
            for (var i = 0; i < 5; i++) {
                var count = plr.mapPlayersOnArea(i);
                pattern += count > 0 ? "1" : "0";
                onPlatforms += count > 0 ? 1 : 0;
            }
            if (onPlatforms !== 3) {
                npc.sendNext("You haven't found the 3 correct platforms yet. Don't forget that you must have 1 person stand in the center of each of the 3 correct platforms to be counted as a correct answer. If necessary, you can place a Platform Puppet to stand in for a character on any platform. Good luck!");
            } else if (pattern === correct) {
                plr.instanceProperty("stage3", 1);
                npc.sendNext("Congratulations on clearing this stage! I will create a portal that will lead you to the next one. You're on a time limit, so please hurry! Good luck!");
            } else {
                var match = 0;
                for (var i = 0; i < 5; i++) if (correct[i] === pattern[i] && correct[i] === "1") match++;
                npc.sendNext("Currently, you've selected " + match + " answer platforms. That is not the correct combination. Keep trying!");
            }
        } else {
            npc.sendBackNext("Hello, Welcome to the 3rd stage. Next to you you'll see barrels with kittens inside on top of the platforms. Out of these platform, #b3 of them lead to the portals for the next stage. 3 of the party members need to find the correct platform to step on and clear the stage. \r\nBUT, you need to stand firm right at the center of it, not standing on the edge, in order to be counted as a correct answer, so make sure to remember that. Also, only 3 members of your party are allowed on the platforms. Once the members are on them, the leader of the party must double-click me to check and see if the answer's right or not#k. Now, find the correct platforms~!", false, true);
            npc.sendBackNext("If there aren't enough people to stand on the platform, purchase a Platform Puppet #v4001454# from Nella and place it on the correct platform. The platform will mistake Platform Puppet for a character. Nifty, huh?", true, true);
            plr.instanceProperty("stage3a", 1);
        }
    }

    else if (stagePart === "4") {
        npc.sendNext("Hello. Welcome to the 4th stage. Walk around the map and you'll be able to find some monsters. The monsters may be familiar to you, but they may be much stronger than you think, so please be careful. Good luck!");
    }

    else if (stagePart === "5") {
        if (plr.mapAllMonsterCount() > 0) {
            npc.sendNext("Hello, welcome to the fifth and final stage. This time, you must defeat the boss, #rKing Sime#k. Good luck!");
        } else {
            var sel = npc.sendMenu("Congratulations! All the stages have been cleared. If you are done, I can lead you outside.", "I want to leave now");
            if (sel === 0) plr.warp(910340000);
        }
    }

} else {

    // ---------- Party member (non-leader) side ----------

    if (stagePart === "1") {
        if (plr.instanceProperty(plr.getName()) === null) {
            npc.sendNext("First, you must complete the mission I give. Once you complete the mission, you will receive a Pass, which will allow you to pass through.");
            npc.sendOk("Your mission is to collect #r" + (Math.floor(Math.random() * 6) + 3) + " coupons#k. You can obtain the coupons by defeating the #rLigators#k found here.");
            plr.instanceProperty(plr.getName(), Math.floor(Math.random() * 6) + 3);
        } else if (plr.instanceProperty(plr.getName()) === "100") {
            npc.sendNext("You've completed the mission! Please help other party members who may have not completed the mission yet.");
        } else {
            var asked = parseInt(plr.instanceProperty(plr.getName()));
            var owned = plr.itemCount(4001007);
            if (asked === owned) {
                npc.sendNext("You've completed the mission! Please help other party members who may have not completed the mission yet.");
                plr.instanceProperty(plr.getName(), 100);
                plr.takeItem(4001007, plr.itemCount(4001007));
                var done = parseInt(plr.instanceProperty("stage1b") || "0") + 1;
                plr.instanceProperty("stage1b", done);
                plr.mapTopMessage("You've completed " + done + " passes.");
                if (done === plr.mapPlayerCount() - 1) {
                    plr.instanceProperty("stage1a", 2);
                    plr.mapMapEffect("All of the individual missions have been cleared. The Party Leader should come talk to me.", 5120017);
                    plr.mapEnvironmentChange("quest/party/clear");
                    plr.mapEnvironmentChange("Party1/Clear");
                }
            } else {
                npc.sendNext("This isn't it! You must bring me the EXACT number of coupons I told you in order to complete the mission. Here, I'll tell you the number again.");
                npc.sendNextPrev("Your mission is to collect #r" + asked + " coupons#k. You can obtain the coupons by defeating the #rLigators#k found here.");
            }
        }
    }

    else if (stagePart === "2") {
        npc.sendNext("Hi. Welcome to the 2nd stage. Next to me, you'll see a number of ropes. Out of these ropes, #b3 are connected to the portal that sends you to the next stage#k. All you need to do is have #b3 party members to find the answer ropes and hang on them#k. \r\nBUT, it doesn't count as an answer if you hang on to the rope too low; please bring yourself up enough to be counted as a correct answer. Also, only 3 members of your party are allowed on the ropes. Once they are hanging on, the leader of the party must #bdouble-click me to check and see if the answer's correct or not#k. Now, find the right ropes to hang on!");
    }

    else if (stagePart === "3") {
        npc.sendNext("Hello, Welcome to the 3rd stage. Next to you you'll see barrels with kittens inside on top of the platforms. Out of these platform, #b3 of them lead to the portals for the next stage. 3 of the party members need to find the correct platform to step on and clear the stage. \r\nBUT, you need to stand firm right at the center of it, not standing on the edge, in order to be counted as a correct answer, so make sure to remember that. Also, only 3 members of your party are allowed on the platforms. Once the members are on them, the leader of the party must double-click me to check and see if the answer's right or not#k. Now, find the correct platforms~!");
    }

    else if (stagePart === "4") {
        npc.sendNext("Hello. Welcome to the 4th stage. Walk around the map and you'll be able to find some monsters. The monsters may be familiar to you, but they may be much stronger than you think, so please be careful. Good luck!");
    }

    else if (stagePart === "5") {
        if (plr.mapAllMonsterCount() > 0) {
            npc.sendNext("Hello, welcome to the fifth and final stage. This time, you must defeat the boss, #rKing Sime#k. Good luck!");
        } else {
            var sel = npc.sendMenu("Congratulations! All the stages have been cleared. If you are done, I can lead you outside.", "I want to leave now");
            if (sel === 0) plr.warp(910340000);
        }
    }
}