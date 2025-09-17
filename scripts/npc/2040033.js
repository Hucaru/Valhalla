// Neru Pet Trainer - stateless flow  
// We're close enough, so we continue directly.  
// Checks for 4031128 are done just-in-time.

if (!cm.haveItem(4031128, 1)) {                     // no letter
    cm.sendOk("My brother told me to take care of the pet obstacle course, but... since I'm so far away from him, I can't help but wanting to goof around ...hehe, since I don't see him in sight, might as well just chill for a few minutes.");
}

cm.sendNext("Eh, that's my brother's letter! Probably scolding me for thinking I'm not working and stuff... Eh? Ahhh... you followed my brother's advice and trained your pet and got up here, huh? Nice!! Since you worked hard to get here, I'll boost your intimacy level with your pet.");

if (cm.haveItem(4031128, 1) && cm.getPlayer().getPet(0) != null) {
    cm.gainItem(4031128, -1);
    var closeArr = [1, 2, 3, 4, 5, 6, 7, 8, 9];
    cm.gainClosenessAll(closeArr[Math.floor(Math.random() * closeArr.length)]);
    cm.sendOk("What do you think? Don't you think you have gotten much closer with your pet? If you have time, train your pet again on this obstacle course... of course, with my brother's permission.");
}

cm.sendNextPrev("Hmmm... did you really get here with your pet? These obstacles are for pets. What are you here for without it?? Get outta here!");