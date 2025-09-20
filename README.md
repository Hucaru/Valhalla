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
- [ ] Allow gm commands to update information displayed at login
- [ ] Propagate character deletion to channels
- [ ] Party sync when channel or world server are restarted
- [ ] Guild sync when channel or world server are restarted

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
- [x] Player use item (scrolls, potions etc)
- [x] Player drop item(s)
- [ ] Player pets
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
- [x] Guild emblem
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
- [x] Communication Window
- [x] Quests
- [x] Reactors
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
    - [OpenMG](https://github.com/sewil/OpenMG)
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

## Kubernetes deployment (optional)

This repository now includes Kubernetes manifests that mirror docker-compose services.

Prerequisites:
- A Kubernetes cluster (minikube, kind, K3s, or cloud)
- kubectl configured to point to the cluster
- A container registry or a way to load local images into your cluster

Build and load the image:
- Build the image from the Dockerfile: `docker build -t valhalla:latest -f Dockerfile .`
- If using kind: `kind load docker-image valhalla:latest`
- If using minikube: `minikube image load valhalla:latest`
- Otherwise, push `valhalla:latest` to a registry your cluster can pull from and update the image in your helm values.

Deploy:
- `helm install valhalla ./helm`

Service discovery changes (compared to docker-compose):
- K8s services use hyphens. Configs inside the pods are adjusted accordingly:
  - login-server, world-server, db

All ports are via ClusterIP, and can be exposed via Ingress-Nginx.

### Exposing Kubernetes Services
You will need to use the `ingress-nginx` deployment to expose your service.

The following values should be used to deploy the helm chart:
```
tcp:
  8484: valhalla/login-server:8484
  8600: valhalla/cashShop-server:8600
  8685: valhalla/channel-server-1:8685
  8684: valhalla/channel-server-2:8684
  8683: valhalla/channel-server-3:8683
... etc 
```
**You will need to add all the channels you intend on having, and the port decreases by 1 for each additional channel**

1. Deploy Ingress-Nginx via Helm
    - `helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx`
    - `helm install -n ingress-nginx ingress-nginx ingress-nginx/ingress-nginx --create-namespace -f values.yaml`
1. Update Maplestory client to use ingress-nginx external IP
    - `kubectl get svc -n ingress-nginx`
1. Update Valhalla config to use ingress-nginx external IP
    - `channel.clientConnectionAddress: "<loadbalancer-ip"`
1. Upgrade/Restart Helm chart
    - `helm upgrade -n valhalla valhalla ./helm -f values.yaml`
