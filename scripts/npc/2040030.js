// 小精靈 - 愛奧斯塔入口
var menu = "What do you want to know? \r\n#L0##bHow do you revive pets?#l\r\n#L1#How do you raise pets?#l\r\n#L2#Can pets die?#l\r\n#L3#Tell me about Action Pets.#l\r\n#L4#How do I change pet stats?#l"
npc.sendSelection(menu)
var select = npc.selection()

switch (select) {
    case 0: // How do you revive pets?
        npc.sendBackNext("I'm #p2040030#, and I research many types magic here in place of my master, #p1032102#. It seems there are a lot of pets in Ludibrium, as well. I'll excuse myself now, as I have plenty of pending tasks to attend to.", true, true)
        break

    case 1: // How do you raise pets?
        npc.sendBackNext("Depending on the command you give, pets can love it, hate, and display other kinds of reactions to it. If you give the pet a command and it follows you well, your intimacy goes up. Double click on the pet and you can check the intimacy, level, fullness and etc...", false, true)
        npc.sendBackNext("Talk to the pet, pay attention to it and its intimacy level will go up and eventually his overall level will go up too. As the intimacy level rises, the pet's overall level will rise soon after. As the overall level rises, one day the pet may even talk like a person a little bit, so try hard raising it. Of course it won't be easy doing so...", true, true)
        npc.sendBackNext("It may be a live doll but they also have life so they can feel the hunger too. #bFullness#k shows the level of hunger the pet's in. 100 is the max, and the lower it gets, it means that the pet is getting hungrier. After a while, it won't even follow your command and be on the offensive, so watch out over that.", true, true)
        npc.sendBackNext("That's right! Pets can't eat the normal human food. Instead a teddy bear in Ludibrium called #b#p2041014##k sells #bPet Food#k so if you need food for your pet, find #b#p2041014##k It'll be a good idea to buy the food in advance and feed the pet before it gets really hungry.", true, true)
        npc.sendBackNext("Oh, and if you don't feed the pet for a long period of time, it goes back home by itself. You can take it out of its home and feed it but it's not really good for the pet's health, so try feeding him on a regular basis so it doesn't go down to that level, alright? I think this will do.", true, true)
        break

    case 2: // Can pets die?
        npc.sendBackNext("Dying... well, they aren't technically ALIVE per se, so I don't know if dying is the right term to use. They are dolls imbued with #p1012005#'s magical power and the Water of Life. Of course while they're animated, they act just like a living animal...", false, true)
        npc.sendBackNext("After enough time has passes, the Water of Life bringing your pets to life will run out and they'll go back to being dolls. But they won't stay like that forever! Just use some Premium Water of Life to revive them!", true, true)
        npc.sendBackNext("Even if it someday moves again, it's sad to see them stop altogether. Please be nice to them while they are alive and moving. Feed them well, too. Isn't it nice to know that there's something alive that follows and listens to only you?", true, true)
        break

    case 3: // Tell me about Action Pets
        npc.sendBackNext("An #baction pet#k is a pet that can transform and evolve. If you use the #r'transform'#k command or #rdefeat Monsters Near Your Level#k, it will take on a new form. The transformed action pet will return to its original form if you enter the #r'return'#k command or wait for it to change back on its own. \r\nAlso, action pets that reach Level 30 can be evolved using the #bAccelerator#k item.", true, true)
        break

    case 4: // How do I change pet stats
        npc.sendBackNext("In order to transfer the pet ability points, closeness and level, Pet AP Reset Scroll is required. If you take this scroll to Mar the Fairy in Ellinia, she will