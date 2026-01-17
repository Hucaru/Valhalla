# Admin Commands

This guide documents all administrator commands available in Valhalla. These commands are accessible to players with GM privileges and are used for server management, player assistance, and debugging.

## Overview

Commands are executed by typing them in the game chat with a forward slash prefix (e.g., `/command`). Most commands support targeting specific players or affecting the command user.

## Command Categories

- [Server Management](#server-management) - Control server-wide settings
- [Player Management](#player-management) - Modify player stats and properties
- [Map & Instance Management](#map--instance-management) - Control map instances
- [Combat & Mobs](#combat--mobs) - Spawn and manage monsters
- [Items & Economy](#items--economy) - Create items and modify currency
- [Quests & Skills](#quests--skills) - Manage quests and player skills
- [Party & Guild](#party--guild) - Party and guild utilities
- [Events](#events) - Event system commands
- [Debugging & Testing](#debugging--testing) - Debug tools and utilities

---

## Server Management

### `/rate <type> <value>`

Changes the server rate multiplier for experience, drops, or mesos.

**Syntax:**
```
/rate <exp | drop | mesos> <rate>
```

**Parameters:**
- `type` - Must be one of: `exp`, `drop`, or `mesos`
- `rate` - Numeric multiplier value (e.g., 2.0 for 2x rate)

**Example:**
```
/rate exp 3.0      # Set 3x EXP rate
/rate drop 2.5     # Set 2.5x drop rate
/rate mesos 1.5    # Set 1.5x mesos rate
```

### `/showRates`

Displays current server rates for EXP, drops, and mesos.

**Example:**
```
/showRates         # Shows: "Exp: x2.00, Drop: x1.50, Mesos: x1.00"
```

### `/setWorldMessage <ribbon> [message]`

Updates the world message displayed at login.

**Syntax:**
```
/setWorldMessage <ribbon_number> [message]
```

**Parameters:**
- `ribbon_number` - Numeric ribbon identifier (0 or greater)
- `message` - Optional message text (omit to clear message)

**Example:**
```
/setWorldMessage 0 Welcome to Valhalla!
/setWorldMessage 1                        # Clears message for ribbon 1
```

### `/header [message]`

Sets or clears the scrolling header message shown to all players on the current channel.

**Example:**
```
/header Server maintenance in 1 hour
/header                                   # Clears the header
```

### `/notice <message>`

Broadcasts a notice message to all players on the channel.

**Example:**
```
/notice Event starting in 5 minutes!
```

### `/msgBox <message>`

Broadcasts a dialogue box message to all players on the channel.

**Example:**
```
/msgBox Please report any bugs to the forums
```

---

## Player Management

### `/hp [player] <amount>`

Sets HP for yourself or a target player.

**Syntax:**
```
/hp <amount>
/hp <player> <amount>
```

**Parameters:**
- `player` - Optional player name
- `amount` - HP value to set

**Notes:**
- If amount exceeds max HP, max HP is also increased

**Example:**
```
/hp 5000           # Set your HP to 5000
/hp Alice 10000    # Set Alice's HP to 10000
```

### `/mp [player] <amount>`

Sets MP for yourself or a target player.

**Syntax:**
```
/mp <amount>
/mp <player> <amount>
```

**Parameters:**
- `player` - Optional player name
- `amount` - MP value to set

**Notes:**
- If amount exceeds max MP, max MP is also increased

**Example:**
```
/mp 3000           # Set your MP to 3000
/mp Bob 8000       # Set Bob's MP to 8000
```

### `/setMaxHP <amount>`

Sets maximum HP for the command user.

**Syntax:**
```
/setMaxHP <amount>
```

**Parameters:**
- `amount` - Maximum HP value (must be at least 1)

**Example:**
```
/setMaxHP 30000    # Set max HP to 30000
```

### `/setMaxMP <amount>`

Sets maximum MP for the command user.

**Syntax:**
```
/setMaxMP <amount>
```

**Parameters:**
- `amount` - Maximum MP value (cannot be negative)

**Example:**
```
/setMaxMP 20000    # Set max MP to 20000
```

### `/str [player] <amount>`

Sets STR stat for yourself or a target player.

**Syntax:**
```
/str <amount>
/str <player> <amount>
```

**Parameters:**
- `player` - Optional player name
- `amount` - STR value (cannot be negative)

**Example:**
```
/str 999           # Set your STR to 999
/str Charlie 500   # Set Charlie's STR to 500
```

### `/dex [player] <amount>`

Sets DEX stat for yourself or a target player.

**Syntax:**
```
/dex <amount>
/dex <player> <amount>
```

**Example:**
```
/dex 999           # Set your DEX to 999
```

### `/int [player] <amount>`

Sets INT stat for yourself or a target player.

**Syntax:**
```
/int <amount>
/int <player> <amount>
```

**Example:**
```
/int 999           # Set your INT to 999
```

### `/luk [player] <amount>`

Sets LUK stat for yourself or a target player.

**Syntax:**
```
/luk <amount>
/luk <player> <amount>
```

**Example:**
```
/luk 999           # Set your LUK to 999
```

### `/exp [player] <amount>`

Sets experience points for yourself or a target player.

**Syntax:**
```
/exp <amount>
/exp <player> <amount>
```

**Example:**
```
/exp 1000000       # Set your EXP to 1,000,000
/exp Diana 500000  # Set Diana's EXP to 500,000
```

### `/gexp [player] <amount>`

Gives experience points (with level-up handling) to yourself or a target player.

**Syntax:**
```
/gexp <amount>
/gexp <player> <amount>
```

**Notes:**
- Triggers level-up effects if enough EXP is gained

**Example:**
```
/gexp 50000        # Give yourself 50,000 EXP
```

### `/ap [player] <amount>`

Sets available ability points for yourself or a target player.

**Syntax:**
```
/ap <amount>
/ap <player> <amount>
```

**Example:**
```
/ap 100            # Set your AP to 100
```

### `/sp [player] <amount>`

Sets available skill points for yourself or a target player.

**Syntax:**
```
/sp <amount>
/sp <player> <amount>
```

**Example:**
```
/sp 50             # Set your SP to 50
```

### `/level [player] <amount>`

Sets the level for yourself or a target player.

**Syntax:**
```
/level <amount>
/level <player> <amount>
```

**Example:**
```
/level 200         # Set your level to 200
/level Eve 100     # Set Eve's level to 100
```

### `/levelup [player] [amount]`

Increases level by the specified amount (default 1).

**Syntax:**
```
/levelup
/levelup <amount>
/levelup <player> <amount>
```

**Example:**
```
/levelup           # Level up once
/levelup 10        # Level up 10 times
```

### `/job [player] <job>`

Changes job for yourself or a target player.

**Syntax:**
```
/job <job_id | job_name>
/job <player> <job_id | job_name>
```

**Parameters:**
- `job` - Can be numeric ID or name (e.g., "Warrior", "FirePoisonMage")

**Supported Job Names:**
- Beginner
- Warrior, Fighter, Crusader, Page, WhiteKnight, Spearman, DragonKnight
- Magician, FirePoisonWizard, FirePoisonMage, IceLightWizard, IceLightMage, Cleric, Priest
- Bowman, Hunter, Ranger, Crossbowman, Sniper
- Thief, Assassin, Hermit, Bandit, ChiefBandit
- Gm, SuperGm

**Example:**
```
/job 110           # Change to Fighter (job ID 110)
/job Priest        # Change to Priest by name
```

### `/kill [player]`

Kills yourself or a target player (sets HP to 0).

**Syntax:**
```
/kill
/kill <player>
```

**Example:**
```
/kill              # Kill yourself
/kill Frank        # Kill Frank
```

### `/revive [player]`

Revives yourself or a target player (restores full HP).

**Syntax:**
```
/revive
/revive <player>
```

**Example:**
```
/revive            # Revive yourself
/revive Grace      # Revive Grace
```

---

## Map & Instance Management

### `/warp [player] <map>`

Warps yourself or a target player to the specified map.

**Syntax:**
```
/warp <map_id | map_name>
/warp <player> <map_id | map_name>
```

**Supported Map Names:**
- `amherst`, `southperry` (Maple Island)
- `lith`, `henesys`, `kerning`, `perion`, `ellinia`, `sleepy` (Victoria Island)
- `orbis`, `elnath`, `ludi`, `omega`, `aqua` (Ossyria)
- `gm`, `balrog`, `guild`, `pap`, `pianus`, `zakum`, `kerningpq`, `ludipq` (Special maps)

**Example:**
```
/warp 100000000    # Warp to Henesys by map ID
/warp henesys      # Warp to Henesys by name
/warp Henry ludi   # Warp Henry to Ludibrium
```

### `/warpTo <player>`

Warps yourself to the location of another player.

**Syntax:**
```
/warpTo <player>
```

**Example:**
```
/warpTo Alice      # Warp to Alice's location
```

### `/whereami`

Shows your current map ID.

**Example:**
```
/whereami          # Displays: "100000000"
```

### `/pos`

Shows your current position coordinates.

**Example:**
```
/pos               # Displays: "(x: 123, y: 456)"
```

### `/mapInfo`

Displays information about all instances in the current map.

**Example:**
```
/mapInfo           # Shows instance details
```

### `/createInstance`

Creates a new instance of the current map.

**Example:**
```
/createInstance    # Creates and returns new instance ID
```

### `/changeInstance [player] <id>`

Changes yourself or a target player to a different instance.

**Syntax:**
```
/changeInstance <instance_id>
/changeInstance <player> <instance_id>
```

**Example:**
```
/changeInstance 2  # Move to instance 2
```

### `/deleteInstance <id>`

Deletes a map instance.

**Syntax:**
```
/deleteInstance <instance_id>
```

**Notes:**
- Cannot delete instance 0
- Cannot delete your current instance

**Example:**
```
/deleteInstance 3  # Delete instance 3
```

### `/clearInstProps`

Clears all properties for the current instance.

**Example:**
```
/clearInstProps
```

### `/properties`

Displays all properties set on the current instance.

**Example:**
```
/properties        # Lists all instance properties
```

### `/enablePortal <name> <true|false>`

Enables or disables a portal by name.

**Syntax:**
```
/enablePortal <portal_name> <true | false>
```

**Example:**
```
/enablePortal exit true   # Enable the 'exit' portal
/enablePortal pq false    # Disable the 'pq' portal
```

### `/removeTimer`

Removes the field timer from the current map instance.

**Example:**
```
/removeTimer
```

---

## Combat & Mobs

### `/spawn <mob_id> [count]`
### `/spawnMob <mob_id> [count]`

Spawns one or more monsters at your location.

**Syntax:**
```
/spawn <mob_id> [count]
```

**Parameters:**
- `mob_id` - Numeric monster ID
- `count` - Optional number to spawn (default: 1)

**Example:**
```
/spawn 9500317     # Spawn 1 Zakum
/spawn 100100 5    # Spawn 5 of mob ID 100100
```

### `/spawnBoss <name> [count]`

Spawns boss monsters by name.

**Syntax:**
```
/spawnBoss <boss_name> [count]
```

**Supported Boss Names:**
- `balrog` - Balrog
- `cbalrog` - Crimson Balrog
- `zakum` - Zakum (spawns all arms + body)
- `pap` - Papulatus
- `pianus` - Pianus
- `mushmom` - Mushmom
- `zmushmom` - Zombie Mushmom

**Example:**
```
/spawnBoss zakum   # Spawn Zakum with all arms
/spawnBoss pap 2   # Spawn 2 Papulatus
```

### `/killMob [spawn_id]`

Kills a specific mob by its spawn ID.

**Syntax:**
```
/killMob <spawn_id>
```

**Example:**
```
/killMob 12345     # Kill mob with spawn ID 12345
```

### `/killAll`
### `/killmobs`

Kills all monsters in the current map instance.

**Example:**
```
/killAll           # Kills all mobs
```

### `/testMob`

Spawns a test mob (Ergoth) at your location.

**Example:**
```
/testMob
```

---

## Items & Economy

### `/item <item_id> [amount]`

Creates an item and adds it to your inventory.

**Syntax:**
```
/item <item_id> [amount]
```

**Parameters:**
- `item_id` - Numeric item ID
- `amount` - Optional quantity (default: 1)

**Example:**
```
/item 2000001      # Get 1 Red Potion
/item 4001126 100  # Get 100 Pig's Ribbons
```

### `/mesos <amount>`

Sets your mesos to the specified amount.

**Syntax:**
```
/mesos <amount>
```

**Example:**
```
/mesos 10000000    # Set mesos to 10 million
```

### `/nx <amount>`

Adds NX to your account.

**Syntax:**
```
/nx <amount>
```

**Example:**
```
/nx 50000          # Add 50,000 NX
```

### `/maplepoints <amount>`

Adds Maple Points to your account.

**Syntax:**
```
/maplepoints <amount>
```

**Example:**
```
/maplepoints 10000 # Add 10,000 Maple Points
```

### `/loadout`

Gives a pre-configured set of endgame equipment and scrolls.

**Items Included:**
- Various perfect endgame weapons for all classes
- Scroll of Protection for Weapon DEF (100x)
- Other useful consumables

**Example:**
```
/loadout
```

### `/clearDrops`

Removes all item drops from the current map instance.

**Example:**
```
/clearDrops
```

### `/drop`

Creates a test drop with various perfect weapons at your location.

**Example:**
```
/drop              # Spawns test drop
```

### `/dropr <drop_id>`

Removes a specific drop by its ID.

**Syntax:**
```
/dropr <drop_id>
```

**Example:**
```
/dropr 54321       # Remove drop with ID 54321
```

---

## Quests & Skills

### `/questFinish <quest_id>`

Completes a quest immediately.

**Syntax:**
```
/questFinish <quest_id>
```

**Example:**
```
/questFinish 1001  # Complete quest 1001
```

### `/questUntil <quest_id> <part>`

Progresses a quest to a specific part.

**Syntax:**
```
/questUntil <quest_id> <part>
```

**Parameters:**
- `quest_id` - Numeric quest ID
- `part` - Progress stage number (0 or greater)

**Example:**
```
/questUntil 1001 3 # Progress quest 1001 to part 3
```

### `/questReset <quest_id>`

Resets a quest completely (removes from completed and in-progress).

**Syntax:**
```
/questReset <quest_id>
```

**Example:**
```
/questReset 1001   # Reset quest 1001
```

### `/skillLv <skill_id> <level|max>`
### `/skillLv <player> <skill_id> <level|max>`

Sets a skill to a specific level or max level.

**Syntax:**
```
/skillLv <skill_id> <level | max>
/skillLv <player> <skill_id> <level | max>
```

**Parameters:**
- `skill_id` - Numeric skill ID
- `level` - Numeric level or "max" for maximum level
- Setting level to 0 removes the skill

**Example:**
```
/skillLv 1001003 max   # Max out Haste skill
/skillLv 1121006 20    # Set White Knight's Charge to level 20
/skillLv 1001003 0     # Remove Haste skill
```

### `/maxSkills`

Sets all available skills across all job classes to their maximum level.

**Example:**
```
/maxSkills         # Max all skills (1000+ skills)
```

### `/resetSkills`

Removes all skills from your character.

**Example:**
```
/resetSkills       # Clear all skills
```

---

## Party & Guild

### `/partyCreate`

Creates a party with you as the leader.

**Example:**
```
/partyCreate
```

### `/guildCreate`

Opens the guild creation interface.

**Example:**
```
/guildCreate
```

### `/guildDisband`

Disbands your current guild.

**Notes:**
- Must be in a guild to use this command

**Example:**
```
/guildDisband
```

### `/guildPoints <amount>`

Adds guild points to your guild.

**Syntax:**
```
/guildPoints <amount>
```

**Notes:**
- Must be in a guild to use this command

**Example:**
```
/guildPoints 50000 # Add 50,000 guild points
```

---

## Events

### `/eventStart <name> <instance_id>`

Starts an event script.

**Syntax:**
```
/eventStart <event_name> <instance_id>
```

**Parameters:**
- `event_name` - Name of the event script
- `instance_id` - Instance ID to run the event in

**Notes:**
- If in a party, all party members in the same instance will participate
- Otherwise, only the command user participates

**Example:**
```
/eventStart ola 1  # Start 'ola' event in instance 1
```

### `/events`

Lists all currently running events with details.

**Example:**
```
/events            # Shows event IDs, participants, and remaining time
```

---

## Debugging & Testing

### `/packet <hex_data>`

Sends a raw packet to test client responses.

**Syntax:**
```
/packet <hex_string>
```

**Parameters:**
- `hex_string` - Hexadecimal packet data (without spaces)

**Example:**
```
/packet 3E00      # Send test packet
```

**Notes:**
- Used for debugging packet structures
- Packets are logged to server console

### `/changeBgm [music_name]`

Changes or clears the background music for the current map instance.

**Syntax:**
```
/changeBgm [music_name]
```

**Example:**
```
/changeBgm Bgm04/ArabPirate  # Change to pirate theme
/changeBgm                    # Clear BGM override
```

### `/wrong`

Plays the "wrong" effect (visual + sound).

**Example:**
```
/wrong             # Shows quest failure effect
```

### `/clear`

Plays the "clear" effect (visual + sound).

**Example:**
```
/clear             # Shows quest completion effect
```

### `/gate`

Shows a gate portal effect.

**Example:**
```
/gate
```

---

## Notes and Best Practices

### Command Targeting

Many commands support optional player targeting:
- `/command <amount>` - Affects yourself
- `/command <player> <amount>` - Affects the named player

### Permission Levels

The command system currently does not implement rank-based permissions, but this is planned for future development:
- **Admin** - Everything, server-wide commands, item generation
- **Game Master** - Bans, channel-wide commands, monster spawning
- **Support** - Player assistance, issue resolution
- **Community** - Event management

Currently, all commands are available to any player with GM status.

### Map and Mob IDs

- Map IDs are typically 9-digit numbers (e.g., 100000000 for Henesys)
- Monster IDs are typically 7-digit numbers (e.g., 9500317 for Zakum)
- Item IDs vary in length depending on item type

### Safety Considerations

- Use `/clearDrops` before spawning many drops to avoid lag
- Be cautious with `/killAll` in boss maps
- Deleting instances with players can cause unexpected behavior
- Setting extremely high stats or levels may cause client issues

### Useful Command Combinations

**Setting up a test character:**
```
/level 200
/maxSkills
/loadout
/ap 999
/sp 999
```

**Quick boss testing:**
```
/clearDrops
/killAll
/spawnBoss zakum
```

**Map instance management:**
```
/mapInfo
/createInstance
/changeInstance 1
```

---

## See Also

- [Configuration Guide](Configuration.md) - Server configuration and settings
- [NPC Scripting](#) - NPC chat and script commands (see main README.md)
- [Event Scripts](#) - Event system documentation
