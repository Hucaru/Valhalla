// Internet-Cafe Premium Road warper
npc.sendNext("Bzzzt~Beep~Boop!! Welcome.. to the Internet Cafe!!.. I.. can warp.. you.. to different.. training areas.. within.. the special.. Internet Cafe.. Premium Road.. areas.... Experience points.. from.. monsters.. are doubled.. as well as.. drops.. and.. mesos!!..")

var sel = npc.sendMenu("Please.. choose.. a.. destination....",
                       "#m190000000#",
                       "#m191000000#",
                       "#m192000000#",
                       "#m195000000#")

plr.warp([190000000, 191000000, 192000000, 195000000][sel])