// Chloe – Pet Lore & Pet AP Reset Scroll (100000200)
npc.sendNext("Hmm... are you raising one of my kids by any chance? I perfected a spell that uses Water of Life to blow life into a doll. People call it the #bPet#k. If you have one with you, feel free to ask me questions.");

var menuSel = npc.sendMenu("What do you want to know more of?",
    "#bTell me more about Pets.",
    "How do I raise Pets?",
    "Do Pets die too?",
    "Please teach me about transferring pet ability points.");

if (menuSel === 0) {
    npc.sendNext("So you want to know more about Pets. Long ago I made a doll, sprayed Water of Life on it, and cast spell on it to create a magical animal. I know it sounds unbelievable, but it's a doll that became an actual living thing. They understand and follow people very well.");
    npc.sendNextPrev("But Water of Life only comes out little at the very bottom of the World Tree, so I can't give him too much time in life ... I know, it's very unfortunate ... but even if it becomes a doll again I can always bring life back into it so be good to it while you're with it.");
    npc.sendNextPrev("Oh yeah, they'll react when you give them special commands. You can scold them, love them ... it all depends on how you take care of them. They are afraid to leave their masters so be nice to them, show them love. They can get sad and lonely fast ...");
} else if (menuSel === 1) {
    npc.sendNext("Depending on the command you give, pets can love it, hate, and display other kinds of reactions to it. If you give the pet a command and it follows you well, your intimacy goes up. Double click on the pet and you can check the intimacy, level, fullness and etc..");
    npc.sendNextPrev("Talk to the pet, pay attention to it and its intimacy level will go up and eventually his overall level will go up too. As the intimacy level rises, the pet's overall level will rise soon after. As the overall level rises, one day the pet may even talk like a person a little bit, so try hard raising it. Of course it won't be easy doing so...");
    npc.sendNextPrev("It may be a live doll but they also have life so they can feel the hunger too. #bFullness#k shows the level of hunger the pet's in. 100 is the max, and the lower it gets, it means that the pet is getting hungrier. After a while, it won't even follow your command and be on the offensive, so watch out over that.");
    npc.sendNextPrev("Oh yes! Pets can't eat the normal human food. Instead my disciple #b#p1012004##k sells #bPet Food#k at the #m100000000# market so if you need food for your pet, find #m100000000#. It'll be a good idea to buy the food in advance and feed the pet before it gets really hungry.");
    npc.sendNextPrev("Oh, and if you don't feed the pet for a long period of time, it goes back home by itself. You can take it out of its home and feed it but it's not really good for the pet's health, so try feeding him on a regular basis so it doesn't go down to that level, alright? I think this will do.");
} else if (menuSel === 2) {
    npc.sendNext("Dying ... well, they aren't technically ALlVE per se, so I don't know if dying is the right term to use. They are dolls with my magical power and the power of Water of Life to become a live object. Of course while it's alive, it's just like a live animal...");
    npc.sendNextPrev("After some time ... that's correct, they stop moving. They just turn back to being a doll, after the effect of magic dies down and Water of Life dries out. But that doesn't mean it's stopped forever, because once you pour Water of Life over, it's going to be back alive.");
    npc.sendNextPrev("Even if it someday moves again, it's sad to see them stop altogether. Please be nice to them while they are alive and moving. Feed them well, too. Isn't it nice to know that there's something alive that follows and listens to only you?");
} else if (menuSel === 3) {
    npc.sendNext("In order to transfer the pet ability points, closeness and level, Pet AP Reset Scroll is required. If you take this scroll to Mar the Fairy in Ellinia, she will transfer the level and closeness of the pet to another one. I am especially giving it to you because I can feel your heart for your pet. However, I can't give this out for free. I can give you this book for 250,000 mesos. Oh, I almost forgot! Even if you have this book, it is no use if you do not have a new pet to tranfer the Ability points.");
    if (npc.sendYesNo("250,000 mesos will be deducted. Do you really want to buy?")) {
        if (plr.mesos() >= 250000 && plr.giveItem(4160011, 1)) {
            plr.giveMesos(-250000);
            npc.sendOk("Thank you. Please be good to your pet.");
        } else {
            npc.sendOk("Please check if your inventory has empty slot or you don't have enough meso.");
        }
    }
}