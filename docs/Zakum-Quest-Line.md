# Zakum Quest Line Documentation

## Overview
The Zakum Quest Line consists of three sequential stages that players must complete before they can challenge Zakum. These stages use custom quest tracking (quest IDs 7000, 7001, 7002) to monitor player progress.

## Quest Stages

### Stage 1: Dead Mine Party Quest (Quest ID 7000)
**Location:** El Nath - Door to Zakum (Map 211042300)  
**NPC:** Adobis (NPC ID 2030008)  
**Type:** Party Quest (requires party)  
**Objective:** Collect keys and exchange them for Fire Ore

**Process:**
1. Party leader talks to Adobis and selects option "Enter the Zakum Party Quest"
2. Adobis checks:
   - Player must be in a party
   - Player must be the party leader
   - At least 1 party member must be on the same map
3. Party is warped into the event instance (Map 280010000)
4. Players have 30 minutes to collect 7 Keys (Item ID 4001016)
5. Players can optionally collect 32 Documents (Item ID 4001015) for bonus Dead Mine Scrolls
6. Players talk to Aura (NPC ID 2032002) at the end
7. Exchange 7 Keys for 1 Fire Ore (Item ID 4031061)
8. Quest 7000 is marked as complete with data "end"
9. Players are warped back to Door to Zakum

**Event Script:** `/scripts/event/zakum_pq.js`  
**Completion NPC:** Aura (NPC ID 2032002) at `/scripts/npc/2032002.js`

### Stage 2: Lava Jump Quest (Quest ID 7001)
**Location:** El Nath - Door to Zakum (Map 211042300)  
**NPC:** Adobis (NPC ID 2030008)  
**Type:** Solo Jump Quest  
**Objective:** Navigate through the jump quest to obtain Breath of Lava

**Process:**
1. Player talks to Adobis and selects option "Enter the Zakum Jump Quest"
2. Player is warped to the Jump Quest map (Map 280020000)
3. Player navigates through the jump quest obstacles
4. Player reaches Lira (NPC ID 2032003) at the end
5. Lira gives:
   - 15,000 EXP
   - 1 Breath of Lava (Item ID 4031062)
6. Quest 7001 is marked as complete with data "end"
7. Player is warped back to Door to Zakum

**Completion NPC:** Lira (NPC ID 2032003) at `/scripts/npc/2032003.js`  
**Exit NPC:** Amon (NPC ID 2030010) at `/scripts/npc/2030010.js` - allows players to exit early

### Stage 3: Item Exchange (Quest ID 7002)
**Location:** El Nath - Door to Zakum (Map 211042300)  
**NPC:** Adobis (NPC ID 2030008)  
**Type:** Item Exchange  
**Objective:** Exchange quest items for Eyes of Fire

**Process:**
1. Player talks to Adobis and selects option "Exchange quest items for Eye of Fire"
2. Player must have:
   - 1 Fire Ore (Item ID 4031061) from Stage 1
   - 1 Breath of Lava (Item ID 4031062) from Stage 2
   - 30 Zombie's Lost Gold Teeth (Item ID 4000082)
3. Items are exchanged for 5 Eyes of Fire (Item ID 4001017)
4. Quest 7002 is marked as complete with data "end"
5. Player can now challenge Zakum

**Handler NPC:** Adobis (NPC ID 2030008) at `/scripts/npc/2030008.js`

## Quest Tracking System

### Custom Quest IDs
- **7000** - Stage 1 (Dead Mine PQ) completion
- **7001** - Stage 2 (Jump Quest) completion  
- **7002** - Stage 3 (Item Exchange) completion

### Quest Data Format
Each quest uses the `setQuestData(questID, "end")` function to mark completion. The data value "end" indicates the stage is complete.

### Checking Progress
Players can see their progress by talking to Adobis. The NPC dialogue shows:
- Stage 1 (Party Quest): [COMPLETE] or [INCOMPLETE]
- Stage 2 (Jump Quest): [COMPLETE] or [INCOMPLETE]
- Stage 3 (Item Exchange): [COMPLETE] or [INCOMPLETE]

### Implementation Details
The custom quests (7000, 7001, 7002) do NOT exist in NX quest data. They are tracked using the `SetQuestData` scripting function which:
1. Creates the quest in the `inProgress` map if it doesn't exist
2. Stores quest data in the database (`character_quests` table)
3. Sends update packets to the client
4. Bypasses NX quest validation (since these are custom tracking quests)

## NPCs

### Adobis (2030008)
Main quest coordinator NPC at the Door to Zakum
- Displays quest progress
- Starts Stage 1 Party Quest event
- Warps players to Stage 2 Jump Quest
- Handles Stage 3 Item Exchange

### Aura (2032002)
Stage 1 completion NPC
- Exchanges Keys for Fire Ore
- Marks Stage 1 as complete
- Optionally exchanges Documents for Dead Mine Scrolls
- Warps players back to Door to Zakum

### Lira (2032003)
Stage 2 completion NPC
- Gives Breath of Lava and EXP
- Marks Stage 2 as complete
- Warps players back to Door to Zakum

### Amon (2030010)
Jump Quest exit NPC
- Allows players to exit the Jump Quest early
- Warps players back to Door to Zakum

## Event System

### Stage 1 Party Quest Event
**Script:** `/scripts/event/zakum_pq.js`

**Features:**
- 30-minute time limit
- Party-based instance system
- Item cleanup on timeout or leaving
- Party leader leaving ends the event for all
- Drops and properties cleared on start

**Event Functions:**
- `start()` - Initializes event, warps party, sets timer
- `beforePortal()` - Portal validation (currently allows all)
- `afterPortal()` - Shows countdown timer after portal use
- `timeout()` - Cleans up items and warps player out
- `playerLeaveEvent()` - Handles early exits and party leader leaving

## Database Schema

The quest data is stored in the `character_quests` table:
- `characterID` - Player ID
- `questID` - Quest ID (7000, 7001, or 7002)
- `record` - Quest data string ("end" for completed)
- `completed` - Boolean (0 for in-progress tracking)
- `completedAt` - Timestamp (0 for tracking quests)

## Configuration

### Map Constants
Defined in `/constant/constants.go`:
- `MapZakumPQ = 280010000` - Zakum Party Quest starting map

### Item IDs
- `4001016` - Keys (Stage 1 PQ drops)
- `4001015` - Documents (Stage 1 PQ optional drops)
- `4031061` - Fire Ore (Stage 1 reward)
- `4031062` - Breath of Lava (Stage 2 reward)
- `4000082` - Zombie's Lost Gold Teeth (Stage 3 requirement)
- `4001017` - Eye of Fire (Stage 3 reward, Zakum entry item)
- `2030007` - Dead Mine Scroll (Stage 1 optional reward)

## Usage

Players must complete all three stages in order to be fully prepared for Zakum:
1. Complete Stage 1 to get Fire Ore
2. Complete Stage 2 to get Breath of Lava  
3. Complete Stage 3 to exchange all items for Eyes of Fire
4. Use Eyes of Fire to enter the Zakum fight

Progress persists across game sessions and is tracked per character.
