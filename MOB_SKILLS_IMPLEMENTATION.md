# Mob Skills and Debuffs Implementation

## Overview
This document describes the implementation of mob skills and debuffs for the Valhalla MapleStory server. The implementation allows mobs to use various skills that can debuff players, buff themselves, or perform special actions.

## Architecture

### Core Components

1. **Monster (`channel/monster.go`)**
   - `performSkill()` method processes mob skill execution
   - Returns skill information for further handling by the life pool
   - Handles MP consumption and skill cooldowns

2. **Life Pool (`channel/pools.go`)**
   - `applyMobDebuffToPlayers()` method applies skill effects to players
   - Handles both player debuffs and mob self-buffs
   - Manages AoE (Area of Effect) skill application

3. **Character Buffs (`channel/buffs.go`)**
   - `AddMobDebuff()` method applies debuffs to individual players
   - Handles proper packet generation for buff/debuff display
   - Manages debuff expiration and removal
   - `DispelAllBuffs()` method for clearing player buffs

4. **Player (`channel/player.go`)**
   - `addMobDebuff()` convenience method for applying mob debuffs
   - Integrates with existing buff system

5. **Handlers (`channel/handlers.go`)**
   - Enhanced mob damage handler to apply debuffs from mob attacks
   - Handles skills that trigger on mob attacks

## Implemented Mob Skills

### Player Debuffs

1. **Seal** (skill.Mob.Seal - ID 120)
   - Prevents players from using skills
   - Uses BuffSeal bit position
   - Duration based on skill data

2. **Darkness** (skill.Mob.Darkness - ID 121)
   - Reduces player visibility/accuracy
   - Uses BuffDarkness bit position
   - Visual effect shown to player

3. **Weakness** (skill.Mob.Weakness - ID 122)
   - Reduces player stats
   - Uses BuffWeakness bit position
   - Stackable effect

4. **Stun** (skill.Mob.Stun - ID 123)
   - Temporarily immobilizes player
   - Uses BuffStun bit position
   - Short duration, high impact

5. **Curse** (skill.Mob.Curse - ID 124)
   - Reduces player stats
   - Uses BuffCurse bit position
   - Medium duration debuff

6. **Poison** (skill.Mob.Poison - ID 125)
   - Damage over time effect
   - Uses BuffPoison bit position
   - Continuous HP drain

7. **Slow** (skill.Mob.Slow - ID 126)
   - Reduces movement speed
   - Uses BuffSpeed bit position with negative value
   - Speed reduction = -(level * 10)

### Special Skills

1. **Dispel** (skill.Mob.Dispel - ID 127)
   - Removes all active buffs from players
   - Does not remove debuffs
   - Instant effect, no duration

2. **HealAoe** (skill.Mob.HealAoe - ID 114)
   - Heals all mobs in the area
   - Heal amount from skill data (skillData.Hp)
   - Affects all mobs in the field instance

### Mob Self-Buffs

1. **Weapon Attack Up** (skill.Mob.WeaponAttackUp - ID 100, WeaponAttackUpAoe - ID 110)
   - Increases physical attack power
   - Sets MobStat.PowerUp flag
   - Prevents duplicate buff application

2. **Magic Attack Up** (skill.Mob.MagicAttackUp - ID 101, MagicAttackUpAoe - ID 111)
   - Increases magic attack power
   - Sets MobStat.MagicUp flag

3. **Weapon Defence Up** (skill.Mob.WeaponDefenceUp - ID 102, WeaponDefenceUpAoe - ID 112)
   - Increases physical defense
   - Sets MobStat.PowerGuardUp flag

4. **Magic Defence Up** (skill.Mob.MagicDefenceUp - ID 103, MagicDefenceUpAoe - ID 113)
   - Increases magic defense
   - Sets MobStat.MagicGuardUp flag

5. **Weapon Immunity** (skill.Mob.WeaponImmunity - ID 140)
   - Makes mob immune to physical attacks
   - Sets MobStat.PhysicalImmune flag

6. **Magic Immunity** (skill.Mob.MagicImmunity - ID 141)
   - Makes mob immune to magic attacks
   - Sets MobStat.MagicImmune flag

## How It Works

### Skill Execution Flow

1. **Client sends move acknowledgment with skill action** (actualAction >= 21 && <= 25)
2. **Monster.performSkill() is called**
   - Validates skill can be used (not sealed)
   - Retrieves skill data from NX files
   - Deducts MP cost
   - Returns skill ID, level, and data if effect should be applied
3. **Pool.applyMobDebuffToPlayers() processes the skill**
   - For debuffs: applies to all players in the field instance
   - For mob buffs: updates the mob's statBuff flags
   - For special skills: executes custom logic
4. **Player.addMobDebuff() applies the effect**
   - Maps mob skill ID to buff bit position
   - Builds packet data with mask, values, and duration
   - Sends packets to player and broadcasts to other players
   - Schedules automatic expiration

### Attack-Based Debuffs

Mobs can also apply debuffs through attacks:
1. Client sends mob damage packet with `mobAttack < -1`
2. Handler reads skill ID and level from packet
3. Retrieves skill data and calculates duration
4. Calls `Player.addMobDebuff()` to apply the effect

### Packet Structure

Debuff packets follow the same structure as player buffs:
- 8-byte mask indicating which stats are affected
- For each set bit: short value, int32 source ID (negative for mob skills), short duration
- Extra byte for combo/charges (not used for mob debuffs)

Foreign buff packets show effects to other players:
- For Darkness, Seal, Weakness: sends int32 skill ID
- Other debuffs may not show to other players

## Design Decisions

### Negative Skill IDs
Mob skills are stored with negative skill IDs (-skillID) to distinguish them from player skills in the buff tracking system.

### Duration Handling
- Skill data stores time in milliseconds
- Converted to seconds for client display: `(ms + 999) / 1000` (rounded up)
- Expiration managed through existing buff timer system

### AoE Application
All player debuffs apply to all players in the field instance. Future enhancements could add:
- Range checking based on mob/player positions
- Limit based on skill data (e.g., `skillData.Limit`)

### Buff Prevention
Mob self-buffs check if the buff is already active before applying to prevent duplicate effects:
```go
if (mob.statBuff & skill.MobStat.PowerUp) > 0 {
    // Already has buff, don't apply again
}
```

## Future Enhancements

### Not Yet Implemented

1. **Seduce** (skill.Mob.Seduce - ID 128)
   - Would reverse player movement controls
   - Requires client-side movement handling

2. **SendToTown** (skill.Mob.SendToTown - ID 129)
   - Teleports player to nearest town
   - Requires map transition logic

3. **PoisonMist** (skill.Mob.PoisonMist - ID 131)
   - Creates persistent AoE poison zone
   - Requires field effect management

4. **CrazySkull** (skill.Mob.CrazySkull - ID 132)
   - Special debuff type
   - Effect unclear from reference

5. **Zombify** (skill.Mob.Zombify - ID 133)
   - Temporary zombie transformation
   - Requires visual effect

6. **Armor Skill** (skill.Mob.ArmorSkill - ID 142)
   - Defense-based effect
   - Implementation unclear

7. **Damage Reflect** (WeaponDamageReflect - ID 143, MagicDamageReflect - ID 144, AnyDamageReflect - ID 145)
   - Reflects damage back to attacker
   - Requires damage calculation integration

8. **Monster Carnival Skills** (McWeaponAttackUp, etc.)
   - Special PvP event skills
   - Requires Monster Carnival system

9. **Summon** (skill.Mob.Summon - ID 200)
   - Spawns additional mobs
   - Requires mob spawning integration

### Potential Improvements

1. **Range-based targeting**
   - Check distance between mob and players
   - Only apply debuffs to players within range

2. **Resistance system**
   - Players could have resistance stats
   - Chance to avoid or reduce debuff duration

3. **Buff stacking**
   - Allow multiple levels of same buff
   - Track buff sources separately

4. **Visual effects**
   - Additional packets for skill animations
   - Mob status indicators

5. **Buff expiration packets**
   - Send mob buff removal packets when buffs expire
   - Update mob visual state

## Testing Recommendations

1. **Debuff Application**
   - Spawn mobs with specific skills
   - Verify debuffs appear on player UI
   - Confirm duration is correct

2. **Debuff Effects**
   - Seal: Try using skills while sealed
   - Slow: Measure movement speed reduction
   - Poison: Verify HP drain over time

3. **Dispel**
   - Apply player buffs
   - Have mob use Dispel
   - Verify all buffs removed

4. **Mob Buffs**
   - Verify mob damage increases with attack buffs
   - Test immunity buffs prevent damage
   - Confirm buffs don't stack

5. **Attack Debuffs**
   - Have mob attack with skill effect
   - Verify debuff applies on hit
   - Test with different mob types

## References

- OpenMG CharacterStatsPacket: https://github.com/sewil/OpenMG/blob/8cd1f461a6efc16ac6cbbd5945a6d88feb35421a/WvsBeta.Game/Packets/CharacterStatsPacket.cs#L153
- OpenMG CharacterBuffs: https://github.com/sewil/OpenMG/blob/8cd1f461a6efc16ac6cbbd5945a6d88feb35421a/WvsBeta.Game/Characters/CharacterBuffs.cs#L171

## Credits

Implementation based on the OpenMG reference code and adapted for the Valhalla server architecture.
