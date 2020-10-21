# Valhalla

[![Actions Status](https://github.com/Hucaru/Valhalla/workflows/Go/badge.svg)](https://github.com/Hucaru/Valhalla/actions)

## What is this?

This project exists to preserve and archive an early version of the game (v28 of global)

## Client modifications

- 00663007 - change to jmp for multiclient
- 0041BD17 - fill with nop to remove internet explorer iframe advert after client close
- 0066520B - push to stack resolution in y
- 00665211 - push to stack resolution in x
- 0066519c - mov 0x0 instead of 0x10 for windowed mode

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
- [x] Server resets login status upon restart for dangling users

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
- [x] Player use skills
- [ ] Player skill logic (haste etc)
- [x] Player inventory (needs a re-write)
- [ ] Player use item (scrolls, potions etc)
- [ ] Player drop item(s)
- [ ] Player pets
- [x] Player stats
- [x] NPC visible
- [x] NPC movement
- [x] NPC basic chat
- [x] NPC shops
- [x] NPC stylist
- [ ] NPC storage
- [ ] PQ scripts
- [ ] Event scripts
- [x] Load scripts from folder (incl. hot loading)
- [x] Map instancing
- [x] Mob visible
- [x] Mob movement
- [ ] Mob attack
- [ ] Mob skills that cause stat changes
- [x] Mob death
- [x] Mob respawn
- [x] Mob spawns mob(s) on death
- [x] Mob drops
- [x] Mob boss HP bar
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
- [x] Server resets login status upon restart for dangling characters
- [ ] Autonomous GM commands which can be started and stopped at will

Metrics:
- [x] Channel population
- [x] Server thread count (OS and Go)
- [x] Server memory usage (heap and stack)
- [ ] Monster kill rate
- [ ] Ongoing trades
- [ ] Ongoing minigames
- [ ] Ongoing npc script interactions

See screenshots section for an example Grafana dashboard

## Acknowledgements

- Sunnyboy for providing a [list](http://forum.ragezone.com/f921/library-idbs-versions-named-addresses-987815/) of idbs for which this project would not have started
- [Vana](https://github.com/retep998/Vana)
- [WvsGlobal](https://github.com/diamondo25/WvsGlobal)
- [NX](https://nxformat.github.io/) file format (see acknowledgements at link)

## NPC chat display info (use this when scripting NPCs)

NPCs are scripted in javscript powered by [goja](https://github.com/dop251/goja)

Taken from [here](http://forum.ragezone.com/f428/add-learning-npcs-start-finish-643364/)

- #b - Blue text.
- #c[itemid]# - Shows how many [itemid] the player has in their inventory.
- #d - Purple text.
- #e - Bold text.
- #f[imagelocation]# - Shows an image inside the .wz files.
- #g - Green text.
- #h # - Shows the name of the player.
- #i[itemid]# - Shows a picture of the item.
- #k - Black text.
- #l - Selection close.
- #m[mapid]# - Shows the name of the map.
- #n - Normal text (removes bold).
- #o[mobid]# - Shows the name of the mob.
- #p[npcid]# - Shows the name of the NPC.
- #q[skillid]# - Shows the name of the skill.
- #r - Red text.
- #s[skillid]# - Shows the image of the skill.
- #t[itemid]# - Shows the name of the item.
- #v[itemid]# - Shows a picture of the item.
- #x - Returns "0%" (need more information on this).
- #z[itemid]# - Shows the name of the item.
- #B[%]# - Shows a 'progress' bar.
- #F[imagelocation]# - Shows an image inside the .wz files.
- #L[number]# Selection open.
- \r\n - Moves down a line.
- \r - Return Carriage
- \n - New Line
- \t - Tab (4 spaces)
- \b - Backwards

## Screenshots

![Bosses](img/bosses.PNG?raw=true "Bosses")

![Metrics](img/metrics.PNG?raw=true "Metrics")