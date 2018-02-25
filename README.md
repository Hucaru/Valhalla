# Valhalla
Golang v28 maplestory server

## Don't use this as most things don't work yet

## Ingame Features Implemented
- All chat
- Other players visible (things like pets not implemented)
- Movement
- Map traversal
- NPC Spawn
- Mob movement (broken movement e.g. mobs jump down ledges in HHG)
- Leveling (inclusing hp, mp increase based on job and stats)
- Stats in character stats can be changed after leveling
- Skill points can be assigned when leveling
- Exp gained shown
- Exp gained causes level up
- Skills used shown to map (some skills dc e.g. gm dragon roar dc's other players if mobs on map?)
- !warp command takes you to different maps if you are an admin extra argument for map pos id can be used
- !packet sends a packet to the client
- !job, !level, !exp, !hp, !mp


## TODO
### Login Server
- Need to change how interserver comms is handled and reduce the number of go routines and channels used with mutexes.
- Need to add on startup to clear loginserver logins, incase of crash and auto-restart.
- Need to do rankings calculation, packet figured out
- Need to figure out what the extra set of equips are in character display. It looks fine but packet structure is odd as it has extra 0xFF seperator

### World Server
Exists and is essentially placeholder.

### Channel Server (not much gameplay added)
- Like login server, need to reduce the number of channels and simplyfy with mutexes
- Refactor and rething internal logic to make it less spagheti like.

## Cash shop
This is at the bottom of the priority list

## Tests
There are none yet

## Usage
### Set-up
### Running

## Screenshot(s)

![Alt text](images/movement.png?raw=true "In Game")
