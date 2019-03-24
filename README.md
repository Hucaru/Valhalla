# Valhalla

## What is this?

This project exists to preserve and archive an early version of the game

## Client modifications

- 00663007 - change to jmp for multiclient
- 0041BD17 - fill with nop to remove internet explorer iframe add after client close
- 0066520B - push to stack resolution in y
- 00665211 - push to stack resolution in x

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

Cashshop server:

Channel server:

- [x] Players can see each other
- [x] Player chat
- [x] GM commands
- [x] Player use skills
- [x] Player exp
- [x] Player level up
- [ ] Player skill logic (haste etc)
- [ ] Player inventory
- [ ] Player use item (scrolls, potions etc)
- [ ] Player pets
- [x] NPC visible
- [x] NPC movement
- [x] NPC basic chat
- [ ] NPC shops
- [ ] NPC stylist
- [ ] NPC storage
- [x] Map instancing
- [x] Mob visible
- [x] Mob movement
- [x] Mob attack
- [ ] Mob skills that cause stat changes
- [x] Mob death
- [x] Mob respawn
- [ ] Mob spawns mob(s) on death
- [ ] Mob drops
- [ ] Trade
- [x] Minigames
- [ ] Communication Window
- [ ] Party
- [ ] Guild
- [ ] Quests
- [ ] Friends list
- [ ] Reactors
- [ ] Whisphers
- [ ] Buddy chat
- [ ] Chat commands (/find etc.)

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
