// Reward NPC for Ludibrium PQ Completion (party2_reward)
// This NPC appears at the bonus stage exit (map 922011100)

var pass = 4001022;
var key = 4001023;

npc.sendOk("Incredible! You've completed all the stages and now you're here enjoying your victory. Wow! My sincere congratulations to each of you for a job well done. Here's a little treat for you. Before accepting, please check that you have an available slot in your use and equip inventories.");

// Check inventory space
if (plr.getUseInventoryFreeSlot() < 1 || plr.getEquipInventoryFreeSlot() < 1) {
    npc.sendOk("Your use and equip inventories must have at least one slot available. Please make the necessary adjustments and then talk to me again.");
} else {
    // Clean up quest items
    if (plr.itemCount(pass) > 0) {
        plr.removeItemsByIDSilent(pass, plr.itemCount(pass));
    }
    if (plr.itemCount(key) > 0) {
        plr.removeItemsByIDSilent(key, plr.itemCount(key));
    }

    // Random reward system based on OpenMG reference
    var rand = Math.floor(Math.random() * 251);
    var rewardID = 0;
    var rewardAmount = 1;

    if (rand == 0) { rewardID = 2000004; rewardAmount = 10; }
    else if (rand == 1) { rewardID = 2000002; rewardAmount = 100; }
    else if (rand == 2) { rewardID = 2000003; rewardAmount = 100; }
    else if (rand == 3) { rewardID = 2000006; rewardAmount = 30; }
    else if (rand == 4) { rewardID = 2022000; rewardAmount = 30; }
    else if (rand == 5) { rewardID = 2022003; rewardAmount = 30; }
    else if (rand == 6) { rewardID = 2040002; rewardAmount = 1; }
    else if (rand == 7) { rewardID = 2040402; rewardAmount = 1; }
    else if (rand == 8) { rewardID = 2040502; rewardAmount = 1; }
    else if (rand == 9) { rewardID = 2040505; rewardAmount = 1; }
    else if (rand == 10) { rewardID = 2040602; rewardAmount = 1; }
    else if (rand == 11) { rewardID = 2040802; rewardAmount = 1; }
    else if (rand == 12) { rewardID = 4003000; rewardAmount = 50; }
    else if (rand == 13) { rewardID = 4010000; rewardAmount = 15; }
    else if (rand == 14) { rewardID = 4010001; rewardAmount = 15; }
    else if (rand == 15) { rewardID = 4010002; rewardAmount = 15; }
    else if (rand == 16) { rewardID = 4010003; rewardAmount = 15; }
    else if (rand == 17) { rewardID = 4010004; rewardAmount = 15; }
    else if (rand == 18) { rewardID = 4010005; rewardAmount = 15; }
    else if (rand == 19) { rewardID = 4010006; rewardAmount = 10; }
    else if (rand == 20) { rewardID = 4020000; rewardAmount = 15; }
    else if (rand == 21) { rewardID = 4020001; rewardAmount = 15; }
    else if (rand == 22) { rewardID = 4020002; rewardAmount = 15; }
    else if (rand == 23) { rewardID = 4020003; rewardAmount = 15; }
    else if (rand == 24) { rewardID = 4020004; rewardAmount = 15; }
    else if (rand == 25) { rewardID = 4020005; rewardAmount = 15; }
    else if (rand == 26) { rewardID = 4020006; rewardAmount = 15; }
    else if (rand == 27) { rewardID = 4020007; rewardAmount = 6; }
    else if (rand == 28) { rewardID = 4020008; rewardAmount = 6; }
    else if (rand == 29) { rewardID = 1032002; rewardAmount = 1; }
    else if (rand == 30) { rewardID = 1032011; rewardAmount = 1; }
    else if (rand == 31) { rewardID = 1032008; rewardAmount = 1; }
    else if (rand == 32) { rewardID = 1102011; rewardAmount = 1; }
    else if (rand == 33) { rewardID = 1102012; rewardAmount = 1; }
    else if (rand == 34) { rewardID = 1102013; rewardAmount = 1; }
    else if (rand == 35) { rewardID = 1102014; rewardAmount = 1; }
    else if (rand == 36) { rewardID = 2040803; rewardAmount = 1; }
    else if (rand == 37) { rewardID = 2070011; rewardAmount = 1; }
    else if (rand == 38) { rewardID = 2043001; rewardAmount = 1; }
    else if (rand == 39) { rewardID = 2043101; rewardAmount = 1; }
    else if (rand == 40) { rewardID = 2043201; rewardAmount = 1; }
    else if (rand == 41) { rewardID = 2043301; rewardAmount = 1; }
    else if (rand == 42) { rewardID = 2043701; rewardAmount = 1; }
    else if (rand == 43) { rewardID = 2043801; rewardAmount = 1; }
    else if (rand == 44) { rewardID = 2044001; rewardAmount = 1; }
    else if (rand == 45) { rewardID = 2044101; rewardAmount = 1; }
    else if (rand == 46) { rewardID = 2044201; rewardAmount = 1; }
    else if (rand == 47) { rewardID = 2044301; rewardAmount = 1; }
    else if (rand == 48) { rewardID = 2044401; rewardAmount = 1; }
    else if (rand == 49) { rewardID = 2044501; rewardAmount = 1; }
    else if (rand == 50) { rewardID = 2044601; rewardAmount = 1; }
    else if (rand == 51) { rewardID = 2044701; rewardAmount = 1; }
    else if (rand >= 96 && rand <= 150) { rewardID = 2000004; rewardAmount = 10; }
    else if (rand >= 151 && rand <= 200) { rewardID = 2000002; rewardAmount = 100; }
    else { rewardID = 2000003; rewardAmount = 100; }

    if (plr.giveItem(rewardID, rewardAmount)) {
        npc.sendOk("You received " + rewardAmount + " #t" + rewardID + "#!");
        plr.warp(221024500); // Exit map
    } else {
        npc.sendOk("Hmmm... are you sure you have a free slot in your use and etc inventories? I cannot reward you for your effort if your inventory is full...");
    }
}
