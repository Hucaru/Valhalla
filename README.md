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
- [ ] Pin
- [x] Display world ribbons
- [x] Display world messages
- [x] Display world status (e.g. overpopulated)
- [x] World selection
- [x] Channel selection
- [x] Create character
- [x] Delete character
- [x] Migrate to channel server
- [x] Show worlds, channels, world status etc from information sent from world server
- [x] Prevent players from accessing dead channel

World server:

- [x] Keep track of player count
- [x] Send information to login server
- [x] Send IP, port to channel for change channel requests
- [ ] Forward whisphers
- [ ] Forward buddy chat
- [ ] Forward party chat
- [ ] Forward guild chat
- [ ] Allow gm command to actiavate exp/drop changes accross all channels
- [ ] Allow gm commands to update information displayed at login

Cashshop server:
- [ ] List items
- [ ] Allow purchases via different currencies

Channel server:

- [x] GM commands
- [x] Players can see each other
- [x] Player can change channel
- [x] Players can see other movement
- [x] Player chat
- [x] player use portal
- [x] Player allocate skill points
- [ ] Player use skills
- [ ] Player skill logic (haste etc)
- [ ] Player inventory
- [ ] Player use item (scrolls, potions etc)
- [ ] Player pets
- [x] Player stats
- [x] NPC visible
- [x] NPC movement
- [ ] NPC basic chat
- [ ] NPC shops
- [ ] NPC stylist
- [ ] NPC storage
- [ ] PQ scripts
- [ ] Event scripts
- [x] Map instancing
- [x] Mob visible
- [x] Mob movement
- [ ] Mob attack
- [ ] Mob skills that cause stat changes
- [ ] Mob death
- [ ] Mob respawn
- [ ] Mob spawns mob(s) on death
- [ ] Mob drops
- [ ] Trade
- [ ] Minigames
- [ ] Communication Window
- [ ] Party
- [ ] Guild
- [ ] Quests
- [ ] Friends list
- [ ] Reactors
- [ ] Whisphers
- [ ] Buddy chat
- [ ] Chat commands (/find etc.)
- [x] Server resets login status upon restart for dangling characters

## TODOs
- Decouple character id from database id

## Acknowledgements

- Sunnyboy for providing a [list](http://forum.ragezone.com/f921/library-idbs-versions-named-addresses-987815/) of idbs for which this project would not have started
- [Vana](https://github.com/retep998/Vana)
- [WvsGlobal](https://github.com/diamondo25/WvsGlobal)
- [NX](https://nxformat.github.io/) file format (see acknowledgements at link)

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