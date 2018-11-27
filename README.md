<p align="center">
  <img src="https://i.imgur.com/mo4tfJF.png"/>
  <br/>Valhalla Logo made with <a href="https://
www.designevo.com/" title="Free Online Logo Maker">DesignEvo</a>
</p>

## What is this?
This project exists to preserve and archive an early version of the game

## Compiling Instructions
Install [docker](https://docs.docker.com/install/) & [docker-compose](https://docs.docker.com/compose/install/). Thats it! 

## Starting the Server

## Features
Login server:
- [x] Login user
- [ ] Pin (might not add this)
- [x] Display world ribbons
- [x] Display world messages
- [x] Display world status (e.g. overpopulated)
- [x] World selection
- [x] Channel selection
- [x] Create character
- [x] Delete character
- [x] Migrate to channel server
- [ ] Show worlds, channels, world status etc from information sent from world server

World server:
- [ ] Keep track of characters in world
- [ ] Send information to login server
- [ ] Send IP, port to channel for change channel requests
- [ ] Forward whisphers
- [ ] Allow gm command to actiavate exp/drop changes accross all channels

Channel server:
- [x] Players can see each other
- [x] Player chat
- [x] GM commands
- [x] Player use skills
- [ ] Player skill logic
- [ ] Player inventory
- [ ] Player pets
- [ ] Player pet items
- [x] NPC visible
- [x] NPC movement
- [x] NPC basic chat
- [ ] NPC shops
- [ ] NPC stylist
- [ ] NPC storage
- [x] Mob visible
- [x] Mob movement
- [x] Mob use skills
- [ ] Mob skill usage effect on players
- [ ] Mob death
- [ ] Mob respawn
- [ ] Mob drops
- [ ] Trade
- [ ] Minigames
- [ ] Communication Window
- [ ] Party
- [ ] Guild
- [ ] Quests
- [ ] Friends list
- [ ] Party quests
- [ ] Whisphers

Cashshop server:

## TODO:
- Redo nx parsing before adding more features that use it

## Acknowledgements 
- Sunnyboy for providing a [list](http://forum.ragezone.com/f921/library-idbs-versions-named-addresses-987815/) of idbs for which this project would not have started
- [Vana](https://github.com/retep998/Vana)
- [WvsGlobal](https://github.com/diamondo25/WvsGlobal)

## NPC chat display info (use this when scripting NPCs)

NPCs are scripted in [anko](https://github.com/mattn/anko)

Taken from [here](http://forum.ragezone.com/f428/add-learning-npcs-start-finish-643364/)
- #b = Blue text.
- #c[itemid]# Shows how many [itemid] the player has in their inventory.
- #d = Purple text.
- #e = Bold text.
- #f[imagelocation]# - Shows an image inside the .wz files.
- #g = Green text.
- #h # - Shows the name of the player.
- #i[itemid]# - Shows a picture of the item.
- #k = Black text.
- #l - Selection close.
- #m[mapid]# - Shows the name of the map.
- #n = Normal text (removes bold).
- #o[mobid]# - Shows the name of the mob.
- #p[npcid]# - Shows the name of the NPC.
- #q[skillid]# - Shows the name of the skill.
- #r = Red text.
- #s[skillid]# - Shows the image of the skill.
- #t[itemid]# - Shows the name of the item.
- #v[itemid]# - Shows a picture of the item.
- #x - Returns "0%" (need more information on this).
- #z[itemid]# - Shows the name of the item.
- #B[%]# - Shows a 'progress' bar.
- #F[imagelocation]# - Shows an image inside the .wz files.
- #L[number]# Selection open.
- \r\n - Moves down a line.
- \r = Return Carriage
- \n = New Line
- \t = Tab (4 spaces)
- \b = Backwards



## Screenshots
