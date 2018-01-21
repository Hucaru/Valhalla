# Valhalla
Practice at writting golang tcp server using the old Maplestory MMORPG client (v28 ~ 2004) as it is well documented and contains a minimal amount of features that would enable me to say the server is complete

## TODO:
- Go through and change all packet write, reads to uint variants
### World Server
- Server sometimes fails to re-connect to dropped login.
- Need to send dropped login previous world id

### Login Server
- Accept pre-registered worlds

### Channel server
- Can get in game with semi static packet
- GM command for sending client packets
- Read in Data.wz
- Set up handlers for various systems e.g maps
- Get packet structures
