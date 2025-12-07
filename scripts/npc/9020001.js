// Cloto

// TODO: Flavour text and align with original text

var mapId = plr.mapID();
var stagePart = mapId % 10;
var bonus = 103000805;

var props = map.properties();

var coupon = 4001007;
var pass = 4001008;

var questions = [{"text": "Sample question", "answer": 1}];
var ropes = ["1110", "1101","1011","0111"];
var cats = ["11100", "11010", "11001", "10110", "10101", "01110", "01101", "01011", "00111"];
var boxes = ["111000", "110010", "110001", "101010", "101001", "100110", "100101", "100011", "010110", "010101", "010011", "001110", "001101", "001011", "000111"];

var rewards = [];

function clear() {
    map.playSound("Party1/Clear");
    map.showEffect("quest/party/clear");
    map.portalEffect("gate");
    props.clear = true
}

function wrong() {
    map.playSound("Party1/Failed");
    map.showEffect("quest/party/wrong_kor");
    props.wrong = true;
}

if (plr.isPartyLeader()) {
    if (stagePart === 0) {
        if (plr.itemCount(pass) >= 3 && !props.clear) {
            props.stage = 2;
        }

        if (props.stage === undefined) {
            npc.sendNext("Hello, and welcome the the first stage. As you can see, this place is full of Ligators. Each Ligator will drop one #bcoupon#k when defeated. Each party member, except the party leader, must come talk to me and then bring me the exact number of #bcoupons#k that I ask for. Once everyone #bcompletes their individual missions#k, the party can move on to the next stage. You must hurry, since the number of stages available depends on how fast you complete this stage. Good luck!");
            props.stage = 1;
        } else if (props.stage === 1) {
            npc.sendNext("I'm sorry, but at least one party member still hasn't completed their mission. Everyone except the part leader must clear their mission to move on.");
        } else if (props.stage === 2) {
            props.stage = 3;
            clear();
            plr.partyGiveExp(10000);
            plr.removeItemsByID(pass, plr.itemCount(pass));
            npc.sendNext("Congratulations on clearing this stage! I will create a portal that will lead you to the next one. You're on a time limit, so please hurry! Good luck!");
        } else {
            npc.sendNext("You all have cleared the quest for this stage. Use the portal to move to the next stage...");
        }
    }
    else if (stagePart === 1) {
        if (props.stage === undefined) {
            npc.sendNext("Hi. Welcome to the 2nd stage. Next to me, you'll see a number of ropes. Out of these ropes, #b3 are connected to the portal that sends you to the next stage#k. All you need to do is have #b3 party members to find the answer ropes and hang on them#k. \r\nBUT, it doesn't count as an answer if you hang on to the rope too low; please bring yourself up enough to be counted as a correct answer. Also, only 3 members of your party are allowed on the ropes. Once they are hanging on, the leader of the party must #bdouble-click me to check and see if the answer's correct or not#k. Now, find the right ropes to hang on!");

            var rand = Math.random();
            var combination = ropes[Math.floor(rand * ropes.length)];

            props.stage = combination;
        } else if (props.clear) {
            npc.sendNext("You all have cleared the quest for this stage. Use the portal to move to the next stage...");
        } else if (props.wrong) { // swallow interrupt retraversal
            props.wrong = false; 
        } else {
            var correct = props.stage;
            var onRopes = 0;
            var pattern = "";

            for (var i = 0; i < 4; i++) {
                var count = map.playersInArea(i);
                pattern += count > 0 ? "1" : "0";
                onRopes += count > 0 ? 1 : 0;
            }

            if (onRopes !== 3) {
                npc.sendNext("It looks like you haven't found the 3 ropes just yet. Please think of a different combination of ropes, Only 3 are allowed to hang on to ropes, and if you hang on to too low, it won't count as an answer; so please keep that in mind. Keep going!");
            } else if (pattern === correct) {
                clear();
                plr.partyGiveExp(10000);
                npc.sendNext("Congratulations on clearing this stage! I will create a portal that will lead you to the next one. You're on a time limit, so please hurry! Good luck!");
            } else {
                wrong();
                npc.sendNext("That is not the correct combination. Keep trying!");
            }
        }
    } else if (stagePart === 2) {
        if (props.stage === undefined) {
            npc.sendNext("Hello, Welcome to the 3rd stage. Next to you you'll see barrels with kittens inside on top of the platforms. Out of these platform, #b3 of them lead to the portals for the next stage. 3 of the party members need to find the correct platform to step on and clear the stage. \r\nBUT, you need to stand firm right at the center of it, not standing on the edge, in order to be counted as a correct answer, so make sure to remember that. Also, only 3 members of your party are allowed on the platforms. Once the members are on them, the leader of the party must double-click me to check and see if the answer's right or not#k. Now, find the correct platforms~!");

            var rand = Math.random();
            var combination = cats[Math.floor(rand * cats.length)];

            props.stage = combination
        } else if (props.clear) {
            npc.sendNext("You all have cleared the quest for this stage. Use the portal to move to the next stage...");
        } else if (props.wrong) { // swallow interrupt retraversal
            props.wrong = false;
        } else {
            var correct = props.stage;
            var onPlatforms = 0;
            var pattern = "";

            for (var i = 0; i < 5; i++) {
                var count = map.playersInArea(i);
                pattern += count > 0 ? "1" : "0";
                onPlatforms += count > 0 ? 1 : 0;
            }
            
            if (onPlatforms !== 3) {
                npc.sendNext("You haven't found the 3 correct platforms yet. Don't forget that you must have 1 person stand in the center of each of the 3 correct platforms to be counted as a correct answer. If necessary, you can place a Platform Puppet to stand in for a character on any platform. Good luck!");
            } else if (pattern === correct) {
                clear();
                plr.partyGiveExp(10000);
                npc.sendNext("Congratulations on clearing this stage! I will create a portal that will lead you to the next one. You're on a time limit, so please hurry! Good luck!");
            } else {
                wrong();
                npc.sendNext("That is not the correct combination. Keep trying!");
            }
        }
    } else if (stagePart === 3) {
        if (props.stage === undefined) {
            npc.sendNext("Hello. Welcome to the 4th stage. <Insert instructions>. Good luck!");

            var rand = Math.random();
            var combination = boxes[Math.floor(rand * boxes.length)];

            props.stage = combination
        } else if (props.clear) {
            npc.sendNext("You all have cleared the quest for this stage. Use the portal to move to the next stage...");
        } else if (props.wrong) { // swallow interrupt retraversal
            props.wrong = false;
        } else {
            var onPlatforms = 0;
            var pattern = "";

            for (var i = 0; i < 6; i++) {
                var count = map.playersInArea(i);
                pattern += count > 0 ? "1" : "0";
                onPlatforms += count > 0 ? 1 : 0;
            }
            if (onPlatforms !== 3) {
                npc.sendNext("You haven't found the 3 correct platforms yet. Don't forget that you must have 1 person stand in the center of each of the 3 correct platforms to be counted as a correct answer. If necessary, you can place a Platform Puppet to stand in for a character on any platform. Good luck!");
            } else if (pattern === props.stage) {
                clear();
                plr.partyGiveExp(10000);
                npc.sendNext("Congratulations on clearing this stage! I will create a portal that will lead you to the next one. You're on a time limit, so please hurry! Good luck!");
            } else {
                wrong();
                npc.sendNext("That is not the correct combination. Keep trying!");
            }
        }
    } else if (stagePart === 4) {
        if (plr.itemCount(pass) >= 10 && !props.clear) {
            props.stage = 2;
        }

        if (props.stage === undefined) {
            npc.sendNext("Hello, welcome to the fifth and final stage. This time, you must defeat the boss, #rKing Slime#k and collect all the monster passes. Good luck!");
            props.stage = 1;
        } else if (props.stage === 1) {
            npc.sendNext("I'm sorry, but you don't have the right number of passes, keep killing monsters to collect them all");
        } else if (props.stage === 2) {
            props.stage = 3;
            clear();
            
            plr.partyGiveExp(10000);
            plr.removeItemsByID(pass, plr.itemCount(pass));

            var sel = npc.sendMenu("Congratulations! All the stages have been cleared. If you are done, I can lead you outside.", "I want to leave now");
            
            if (sel === 0) {
                // TODO: Give item reward
                plr.warp(bonus);
            }
        } else {
            var sel = npc.sendMenu("Congratulations! All the stages have been cleared. If you are done, I can lead you outside.", "I want to leave now");
            
            if (sel === 0) {
                plr.warp(bonus);
            }
        }
    }
} else {
    if (stagePart === 0) {
        if (props[plr.name()] === undefined) {
            var index = 0;
            npc.sendOk(questions[index].text);
            props[plr.name()] = questions[index];
        } else if (props[plr.name()].answer <= plr.itemCount(coupon) && !props[plr.name()].finished) {
            npc.sendOk("That's correct! Please hand your pass to the party leader.");
            plr.giveItem(pass, 1);
            props[plr.name()].finished = true
        } else if (props[plr.name()].finished) {
            npc.sendOk("You have finished this stage");
        } else {
            npc.sendOk(props[plr.name()].text);
        }
    } else if (stagePart === 1) {
        npc.sendNext("Hi. Welcome to the 2nd stage. Next to me, you'll see a number of ropes. Out of these ropes, #b3 are connected to the portal that sends you to the next stage#k. All you need to do is have #b3 party members to find the answer ropes and hang on them#k. \r\nBUT, it doesn't count as an answer if you hang on to the rope too low; please bring yourself up enough to be counted as a correct answer. Also, only 3 members of your party are allowed on the ropes. Once they are hanging on, the leader of the party must #bdouble-click me to check and see if the answer's correct or not#k. Now, find the right ropes to hang on!");
    } else if (stagePart === 2) {
        npc.sendNext("Hello, Welcome to the 3rd stage. Next to you you'll see barrels with kittens inside on top of the platforms. Out of these platform, #b3 of them lead to the portals for the next stage. 3 of the party members need to find the correct platform to step on and clear the stage. \r\nBUT, you need to stand firm right at the center of it, not standing on the edge, in order to be counted as a correct answer, so make sure to remember that. Also, only 3 members of your party are allowed on the platforms. Once the members are on them, the leader of the party must double-click me to check and see if the answer's right or not#k. Now, find the correct platforms~!");
    } else if (stagePart === 3) {
        npc.sendNext("<instructions>");
    } else if (stagePart === 4) {
        if (!props.clear) {
            npc.sendNext("Hello, welcome to the fifth and final stage. This time, you must defeat the boss, #rKing Slime#k and collect all the monster passes. Good luck!");
        } else {
            var sel = npc.sendMenu("Congratulations! All the stages have been cleared. If you are done, I can lead you outside.", "I want to leave now");

            if (sel === 0) {
                // TODO: Give item reward
                plr.warp(bonus);
            }
        }
    }
}