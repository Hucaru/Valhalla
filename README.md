# Valhalla
Golang v28 maplestory server

## TODO:
- Go through and change all packet write, reads to uint variants where appropriate
### World Server
- Server sometimes fails to re-connect to dropped login.
- Need to send dropped login previous world id

### Login Server
- Accept pre-registered worlds

### Channel server
- Can get in game:
  - Equips
  - Cash equips
  - Inventory equips
  - No items slots 2 - 5 (currently static slot 2, partially understood)
  - No Skills
  - No Quests
- GM command for sending client packets
- Read in Data.nx
- Set up handlers for various systems e.g maps
- Keep probing opcodes for structure

![Alt text](images/server_select.png?raw=true "Server Select")
![Alt text](images/character_select.png?raw=true "Character Select")
![Alt text](images/ingame.png?raw=true "In Game")
