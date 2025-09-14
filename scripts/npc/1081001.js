// So you want to leave Florina Beach?
npc.sendNext("So you want to leave #b#m110000000##k? If you want, I can take you back to #bLith Harbor#k.")

if (npc.sendYesNo("Are you sure you want to return to #b#m104000000##k?")) {
    plr.warp(104000000)
} else {
    npc.sendOk("You must have some business to take care of here. It's not a bad idea to take some rest at #b#m110000000##k! Look at me; I love it here so much that I wound up living here. Hahaha anyway, talk to me when you feel like going back.")
}