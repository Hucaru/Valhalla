// Lady Hikari: bathhouse entry
if (npc.sendYesNo("Would you like to enter the bathhouse? That'll be 300 mesos for you. And don't take the towels!")) {
    if (plr.mesos() < 300) {
        npc.sendOk("Please check your wallet or purse and see if you have 300 mesos to enter this place. We have to keep the water hot, you know...");
    } else {
        plr.takeMesos(300);
        const targetMap = 801000100 + 100 * plr.getGender();
        plr.warp(targetMap, 2); // warp to portal 2
    }
} else {
    npc.sendOk("Please come back some other time.");
}