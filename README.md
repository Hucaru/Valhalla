![Alt text](img/logo.png?raw=true "Valhalla")

[![Actions Status](https://github.com/Hucaru/Valhalla/workflows/Go/badge.svg)](https://github.com/Hucaru/Valhalla/actions)
[Visit our Discord channel](https://discord.gg/KHky9Qy9jF)
## What is this?

This project exists to preserve and archive an early version of the game (v28 of global)

## Client modifications

A DLL which will auto hook the functions to make a localhost and window mode can be found [here](https://github.com/Hucaru/maplestory-client-hook)

## Features

General:
- [x] Simulated latency with jitter (set in config) to make dev environment simulate a real world connections when within network

Login server:
- [x] Login user
- [x] Auto-register (optional feature to automatically create accounts on first login attempt)
- [x] Show EULA on first login
- [X] Pin
- [x] Display world ribbons
- [x] Display world messages
- [x] Display world status (e.g. overpopulated)
- [x] World selection
- [x] Channel selection
- [x] Create character
- [x] Delete character
- [x] Delete character informs world servers
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
- [x] Allow gm commands to update information displayed at login
- [x] Propagate character deletion to channels
- [x] Party sync when channel or world server are restarted
- [x] Guild sync when channel or world server are restarted

Cashshop server:
- [x] List items
- [x] Allow purchases via different currencies

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
- [x] Player skill logic (haste etc)
- [x] Player inventory (needs a re-write)
- [x] Player use item (scrolls, potions etc
- [x] Player use cash item (super megaphones, etc)
- [x] Player drop item(s)
- [x] Player pets
- [x] NPC visible
- [x] NPC movement
- [x] NPC basic chat
- [x] NPC shops
- [x] NPC stylist
- [x] NPC storage
- [x] Load scripts from folder (incl. hot loading)
- [x] Map instancing
- [x] Mob visible
- [x] Mob movement
- [x] Mob attack
- [x] Mob skills that cause stat changes or summon other mobs (not on death)
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
- [x] Guild emblem
- [x] Guild chat
- [x] Guild points update
- [x] Guild rank titles change
- [x] Guild rank update
- [x] Guild notice update
- [x] Guild expel
- [x] Guild member online notice
- [ ] Kerning PQ
- [ ] Ludi PQ
- [x] Balrog boat invasion
- [x] Deleted character removes from guild
- [x] Deleted character removes from party
- [x] Trade
- [x] Communication Window
- [x] Quests
- [x] Reactors
- [x] Server resets login status upon restart for dangling characters

Metrics:
- [x] Channel population
- [x] Server thread count (OS and Go)
- [x] Server memory usage (heap and stack)
- [x] Monster kill rate
- [x] Ongoing trades
- [x] Ongoing minigames
- [x] Ongoing npc script interactions
- [x] Number of parties

See screenshots section for an example Grafana dashboard

## Acknowledgements

- Sunnyboy for providing a [list](http://forum.ragezone.com/f921/library-idbs-versions-named-addresses-987815/) of idbs for which this project would not have started
- The following projects were used to help reverse packet structures that were not clearly shown in the idb
    - [Vana](https://github.com/retep998/Vana)
    - [WvsGlobal](https://github.com/diamondo25/WvsGlobal)
    - [OpenMG](https://github.com/sewil/OpenMG)
- [NX](https://nxformat.github.io/) file format (see acknowledgements at link)

## Getting Started

Valhalla supports multiple deployment methods. Choose the one that best fits your needs:

📚 **[Installation Guide](docs/Installation.md)** - Start here! Covers Data.wz conversion and client setup

### Quick Links by Deployment Method

- 🖥️ **[Local Setup](docs/Local.md)** - Run directly on your machine (best for quick testing)
- 🐳 **[Docker Setup](docs/Docker.md)** - Run using Docker Compose (recommended for most users)
- ☸️ **[Kubernetes Setup](docs/Kubernetes.md)** - Deploy to a Kubernetes cluster (for production)
- 🔨 **[Building from Source](docs/Building.md)** - Build for development work

### Configuration

⚙️ **[Configuration Guide](docs/Configuration.md)** - Complete reference for all configuration options

All server types support both TOML configuration files and environment variables. See the Configuration Guide for details on:
- Command line flags (`-type`, `-config`, `-metrics-port`)
- Database settings
- Server-specific options (login, world, channel, cashshop)
- Network configuration
- Performance tuning

## Advanced Topics

### NPC Scripting

NPCs are scripted in JavaScript powered by [goja](https://github.com/dop251/goja). For detailed NPC chat formatting codes and scripting information, see the scripts directory and existing NPC implementations.

For NPC chat display formatting reference, see the [NPC Chat Formatting](#npc-chat-formatting) section below.

### Production Deployments

- **[Kubernetes](docs/Kubernetes.md)** - Production-ready deployment with Helm, ingress, scaling, and monitoring
- **[Docker](docs/Docker.md)** - Containerized deployment with Docker Compose

## NPC Chat Formatting

When scripting NPCs in JavaScript, use these formatting codes:

- `#b` - Blue text
- `#c[itemid]#` - Shows how many [itemid] the player has in inventory
- `#d` - Purple text
- `#e` - Bold text
- `#f[imagelocation]#` - Shows an image from .wz files
- `#g` - Green text
- `#h #` - Shows the player's name
- `#i[itemid]#` - Shows a picture of the item
- `#k` - Black text
- `#l` - Selection close
- `#m[mapid]#` - Shows the name of the map
- `#n` - Normal text (removes bold)
- `#o[mobid]#` - Shows the name of the mob
- `#p[npcid]#` - Shows the name of the NPC
- `#q[skillid]#` - Shows the name of the skill
- `#r` - Red text
- `#s[skillid]#` - Shows the image of the skill
- `#t[itemid]#` - Shows the name of the item
- `#v[itemid]#` - Shows a picture of the item
- `#x` - Returns "0%" (usage varies)
- `#z[itemid]#` - Shows the name of the item
- `#B[%]#` - Shows a progress bar
- `#F[imagelocation]#` - Shows an image from .wz files
- `#L[number]#` - Selection open
- `\r\n` - Moves down a line
- `\r` - Return carriage
- `\n` - New line
- `\t` - Tab (4 spaces)
- `\b` - Backwards

Reference from [RageZone forums](http://forum.ragezone.com/f428/add-learning-npcs-start-finish-643364/)
