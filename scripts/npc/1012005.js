npc.sendBackNext("Hmm... are you raising one of my kids by any chance? I perfected a spell that uses Water of Life to blow life into a doll. People call it the #bPet#k. If you have one with you, feel free to ask me questions.", false, true)

var menu = "What do you want to know more of? \r\n#L0##bTell me more about Pets.#l\r\n#L1#How do I raise Pets?#l\r\n#L2#Do Pets die too?#l\r\n#L3#Please teach me about transferring pet ability points.#l"
npc.sendSelection(menu)
var select = npc.selection()

if (select == 0) {
    npc.sendBackNext("So you want to know more about Pets. Long ago I made a doll, sprayed Water of Life on it, and cast spell on it to create a magical animal. I know it sounds unbelievable, but it's a doll that became an actual living thing. They understand and follow people very well.", false, true)
    npc.sendBackNext("But Water of Life only comes out little at the very bottom of the World Tree, so I can't give him too much time in life ... I know, it's very unfortunate ... but even if it becomes a doll again I can always bring life back into it so be good to it while you're with it.", true, true)
    npc.sendBackNext("Oh yeah, they'll react when you give them special commands. You can scold them, love them ... it all depends on how you take care of them. They are afraid to leave their masters so be nice to them, show them love. They can get sad and lonely fast ...", true, false)
} else if (select == 1) {
    npc.sendBackNext("Depending on the command you give, pets can love it, hate, and display other kinds of reactions to it. If you give the pet a command and it follows you well, your intimacy goes up. Double click on the pet and you can check the intimacy, level, fullness and etc..", false, true)
    npc.sendBackNext("Talk to the pet, pay attention to it and its intimacy level will go up and eventually his overall level will go up too. As the intimacy level rises, the pet's overall level will rise soon after. As the overall level rises, one day the pet may even talk like a person a little bit, so try hard raising it. Of course it won't be easy doing so...", true, true)
    npc.sendBackNext("It may be a live doll but they also have life so they can feel the hunger too. #bFullness#k shows the level of hunger the pet's in. 100 is the max, and the lower it gets, it means that the pet is getting hungrier. After a while, it won't even follow your command and be on the offensive, so watch out over that.", true, true)
    npc.sendBackNext("Oh yes! Pets can't eat the normal human food. Instead my disciple #b#p1012004##k sells #bPet Food#k at the #m100000000# market so if you need food for your pet, find #m100000000#. It'll be a good idea to buy the food in advance and feed the pet before it gets really hungry.", true, true)
    npc.sendBackNext("Oh, and if you don't feed the pet for a long period of time, it goes back home by itself. You can take it out of its home and feed it but it's not really good for the pet's health, so try feeding him on a regular basis so it doesn't go down to that level, alright? I think this will do.", true, false)
} else if (select == 2) {
    npc.sendBackNext("Dying ... well, they aren't technically ALlVE per se, so I don't know if dying is the right term to use. They are dolls with my magical power and the power of Water of Life to become a live object. Of course while it's alive, it's just like a live animal...", false, true)
    npc.sendBackNext("After some time ... that's correct, they stop moving. They just turn back to being a doll, after the effect of magic dies down and Water of Life dries out. But that doesn't mean it's stopped forever, because once you pour Water of Life over, it's going to be back alive.", true, true)
    npc.sendBackNext("Even if it someday moves again, it's sad to see them stop altogether. Please be nice to them while they are alive and moving. Feed them well, too. Isn't it nice to know that there's something alive that follows and listens to only you?", true, false)
} else if (select == 3) {
    npc