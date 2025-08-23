![Alt text](img/logo.png?raw=true "Valhalla")

[![Actions Status](https://github.com/Hucaru/Valhalla/workflows/Go/badge.svg)](https://github.com/Hucaru/Valhalla/actions)

## What is this?

This project exists to preserve and archive an early version of the game (v28 of global)

## Client modifications

A DLL which will auto hook the functions to make a localhost and window mode can be found [here](https://github.com/Hucaru/maplestory-client-hook)

## Features

General:
- [x] Simulated latency with jitter (set in config) to make dev environment simulate a real world connections when within network

Login server:
- [x] Login user
- [ ] Show EULA on first login
- [ ] Perform gender select on first login
- [X] Pin
- [x] Display world ribbons
- [x] Display world messages
- [x] Display world status (e.g. overpopulated)
- [x] World selection
- [ ] Lock character creation on world if full
- [x] Channel selection
- [x] Create character
- [x] Delete character
- [ ] Delete character informs world server
- [x] Migrate to channel server
- [x] Show worlds, channels, world status etc from information sent from world server
- [x] Prevent players from accessing dead channel
- [x] Server resets login status upon restart for dangling users

World server:
- [x] Keep track of player count
- [x] Send information to login server
- [x] Send IP, port to channel for change channel requests
- [x] Forward player connects to channels
- [x] Forward player leaves game to channels
- [x] Broadcast buddy events
- [x] Broadcast party events
- [x] Broadcast guild events
- [x] Forward whisphers
- [x] Allow gm command to actiavate exp/drop changes accross all channels
- [ ] Allow gm commands to update information displayed at login
- [ ] Propagate character deletion to channels
- [ ] Party sync when channel or world server are restarted
- [ ] Guild sync when channel or world server are restarted

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
- [x] Player stats
- [x] Player use skills
- [ ] Player skill logic (haste etc)
- [x] Player inventory
- [ ] Player use item (scrolls, potions etc)
- [ ] Player drop item(s)
- [ ] Player pets
- [x] NPC visible
- [x] NPC movement
- [x] NPC basic chat
- [x] NPC shops
- [x] NPC stylist
- [ ] NPC storage
- [x] Load scripts from folder (incl. hot loading)
- [x] Map instancing
- [x] Mob visible
- [x] Mob movement
- [x] Mob attack
- [ ] Mob skills that cause stat changes or summon other mobs (not on death)
- [x] Mob death
- [x] Mob respawn
- [x] Mob spawns mob(s) on death
- [x] Mob drops
- [x] Mob boss HP bar
- [x] Minigames
- [x] Whisphers
- [x] Find / Map in buddy window
- [x] Buddy list
- [x] Buddy chat
- [x] Party creation
- [x] Party invite
- [x] Party accept/reject
- [x] Party expel
- [x] Party chat
- [ ] Party HP bar
- [x] Guild creation/disband
- [x] Guild invite
- [x] Guild join/leave
- [x] Guild emblem (bug where server needs restart for it to work properly)
- [x] Guild chat
- [x] Guild points update
- [x] Guild rank titles change
- [x] Guild rank update
- [x] Guild notice update
- [x] Guild expel
- [x] Guild member online notice
- [ ] PQs
- [ ] Guild PQ
- [x] Balrog boat invasion
- [ ] Deleted character removes from guild
- [ ] Deleted character removes from party
- [ ] Trade
- [ ] Communication Window
- [ ] Quests
- [ ] Reactors
- [x] Server resets login status upon restart for dangling characters

Metrics:
- [x] Channel population
- [x] Server thread count (OS and Go)
- [x] Server memory usage (heap and stack)
- [ ] Monster kill rate
- [ ] Ongoing trades
- [ ] Ongoing minigames
- [ ] Ongoing npc script interactions
- [ ] Number of parties

See screenshots section for an example Grafana dashboard

## TODOs

- Profile the channel server and do the following:
    - Reduce branches in frequent paths
    - Determine which pieces of data if any provide any benefit in being converted SOAs
- Implement AES crypt (ontop of the shanda) and determine how to enable it in the client
- Move player save database operations into relevant systems
- Player inventory needs a re-write
- Investigate party reject invite packet from client (it looks like garbage)

## Acknowledgements

- Sunnyboy for providing a [list](http://forum.ragezone.com/f921/library-idbs-versions-named-addresses-987815/) of idbs for which this project would not have started
- The following projects were used to help reverse packet structures that were not clearly shown in the idb
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
