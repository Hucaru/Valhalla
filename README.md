# Valhalla
Golang v28 maplestory server

## Don't use this as most things don't work yet

## Login Server
- Need to change how interserver comms is handled and reduce the number of go routines and channels used with mutexes.
- Need to add on startup to clear loginserver logins, incase of crash and auto-restart.
- Need to do rankings
- Need to figure out what the extra set of equips are in character display. It looks fine but packet structure is odd as it has extra 0xFF seperator

## World Server
Exists and is essentially placeholder.

## Channel Server (not much gameplay added)
- Parses nx file. 
- Gets ingame and loads npcs. 
- !warp command takes you to different maps if you are an admin.
- All map chat works.
- Movement works
- Like login server, need to reduce the number of channels and simplyfy with mutexes

## Cash shop
This is at the bottom of the priority list

## Tests
There are none yet

## Usage
### Set-up
### Running

## Screenshot(s)

![Alt text](images/movement.png?raw=true "In Game")
