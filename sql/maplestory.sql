-- Adminer 5.3.0 MySQL 5.7.44 dump

SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `accounts`;
CREATE TABLE `accounts` (
  `accountID` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` tinytext NOT NULL,
  `password` tinytext NOT NULL,
  `pin` tinytext NOT NULL,
  `isLogedIn` tinyint(4) NOT NULL DEFAULT '0',
  `adminLevel` tinyint(4) NOT NULL DEFAULT '0',
  `isBanned` int(11) NOT NULL DEFAULT '0',
  `gender` tinyint(4) NOT NULL DEFAULT '0',
  `dob` int(11) NOT NULL,
  `eula` tinyint(4) NOT NULL,
  `nx` int(11) unsigned NOT NULL DEFAULT '0',
  `maplepoints` int(11) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`accountID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


DROP TABLE IF EXISTS `buddy`;
CREATE TABLE `buddy` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `characterID` int(11) NOT NULL,
  `friendID` int(11) NOT NULL,
  `accepted` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 is accepted, 1 is request pending',
  PRIMARY KEY (`id`),
  KEY `characterID` (`characterID`),
  KEY `friendID` (`friendID`),
  CONSTRAINT `buddy_ibfk_4` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`) ON DELETE CASCADE,
  CONSTRAINT `buddy_ibfk_5` FOREIGN KEY (`friendID`) REFERENCES `characters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


DROP TABLE IF EXISTS `characters`;
CREATE TABLE `characters` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `accountID` int(10) unsigned NOT NULL,
  `guildID` int(11) DEFAULT NULL,
  `guildRank` tinyint(4) NOT NULL DEFAULT '1',
  `worldID` int(11) unsigned NOT NULL,
  `channelID` tinyint(2) NOT NULL DEFAULT '-1',
  `previousChannelID` tinyint(2) NOT NULL DEFAULT '-1',
  `migrationID` tinyint(4) NOT NULL DEFAULT '-1',
  `name` tinytext NOT NULL,
  `gender` int(11) unsigned NOT NULL,
  `skin` int(11) unsigned NOT NULL,
  `hair` int(11) unsigned NOT NULL,
  `face` int(11) unsigned NOT NULL,
  `level` int(200) unsigned NOT NULL DEFAULT '1',
  `job` int(11) unsigned NOT NULL DEFAULT '0',
  `str` int(11) unsigned NOT NULL,
  `dex` int(11) unsigned NOT NULL,
  `intt` int(11) unsigned NOT NULL,
  `luk` int(11) unsigned NOT NULL,
  `hp` int(11) unsigned NOT NULL DEFAULT '100',
  `maxHP` int(11) unsigned NOT NULL DEFAULT '100',
  `mp` int(11) unsigned NOT NULL DEFAULT '50',
  `maxMP` int(11) unsigned NOT NULL DEFAULT '50',
  `ap` int(11) unsigned NOT NULL DEFAULT '0',
  `sp` int(11) unsigned NOT NULL DEFAULT '0',
  `exp` int(11) unsigned NOT NULL DEFAULT '0',
  `fame` int(11) NOT NULL DEFAULT '0',
  `mapID` int(11) unsigned NOT NULL DEFAULT '0',
  `mapPos` int(11) unsigned NOT NULL DEFAULT '0',
  `previousMapID` int(11) unsigned NOT NULL DEFAULT '0',
  `mesos` int(11) unsigned NOT NULL DEFAULT '0',
  `equipSlotSize` tinyint(4) NOT NULL DEFAULT '32',
  `useSlotSize` tinyint(4) NOT NULL DEFAULT '32',
  `setupSlotSize` tinyint(4) NOT NULL DEFAULT '32',
  `etcSlotSize` tinyint(4) NOT NULL DEFAULT '32',
  `cashSlotSize` tinyint(4) NOT NULL DEFAULT '32',
  `miniGameWins` int(11) NOT NULL DEFAULT '0',
  `miniGameDraw` int(11) NOT NULL DEFAULT '0',
  `miniGameLoss` int(11) NOT NULL DEFAULT '0',
  `miniGamePoints` int(11) NOT NULL DEFAULT '2000',
  `buddyListSize` tinyint(3) unsigned NOT NULL DEFAULT '20',
  `inCashShop` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `userID` (`accountID`),
  KEY `guildID` (`guildID`),
  CONSTRAINT `characters_ibfk_2` FOREIGN KEY (`accountID`) REFERENCES `accounts` (`accountID`) ON DELETE CASCADE,
  CONSTRAINT `characters_ibfk_4` FOREIGN KEY (`guildID`) REFERENCES `guilds` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


DROP TABLE IF EXISTS `guilds`;
CREATE TABLE `guilds` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `worldID` int(11) NOT NULL,
  `capacity` int(11) NOT NULL DEFAULT '50',
  `name` tinytext NOT NULL,
  `notice` text NOT NULL,
  `master` tinytext NOT NULL,
  `jrMaster` tinytext NOT NULL,
  `member1` tinytext NOT NULL,
  `member2` tinytext NOT NULL,
  `member3` tinytext NOT NULL,
  `logoBg` smallint(6) NOT NULL DEFAULT '0',
  `logoBgColour` smallint(6) NOT NULL DEFAULT '0',
  `logo` smallint(6) NOT NULL DEFAULT '0',
  `logoColour` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `points` int(10) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

DROP TABLE IF EXISTS `guild_invites`;
CREATE TABLE `guild_invites` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `playerID` int(11) NOT NULL,
  `guildID` int(11) NOT NULL,
  `inviter` tinytext NOT NULL,
  PRIMARY KEY (`id`),
  KEY `playerID` (`playerID`),
  KEY `guildID` (`guildID`),
  CONSTRAINT `guild_invites_ibfk_3` FOREIGN KEY (`playerID`) REFERENCES `characters` (`id`) ON DELETE CASCADE,
  CONSTRAINT `guild_invites_ibfk_4` FOREIGN KEY (`guildID`) REFERENCES `guilds` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

DROP TABLE IF EXISTS `character_buffs`;
CREATE TABLE IF NOT EXISTS character_buffs (
   `characterID` INT NOT NULL,
   `sourceID` INT NOT NULL,
   `level` TINYINT NOT NULL,
   `expiresAtMs` BIGINT NOT NULL,
   PRIMARY KEY(`characterID`, `sourceID`),
   CONSTRAINT `buffs_ibfk_5` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

DROP TABLE IF EXISTS `items`;
CREATE TABLE `items` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `characterID` int(11) NOT NULL,
  `itemID` int(11) NOT NULL,
  `inventoryID` int(11) NOT NULL DEFAULT '1',
  `slotNumber` int(11) NOT NULL,
  `amount` int(11) NOT NULL DEFAULT '1',
  `flag` tinyint(4) NOT NULL DEFAULT '0',
  `upgradeSlots` tinyint(4) NOT NULL DEFAULT '0',
  `level` tinyint(4) NOT NULL DEFAULT '0',
  `str` smallint(6) NOT NULL DEFAULT '0',
  `dex` smallint(6) NOT NULL DEFAULT '0',
  `intt` smallint(6) NOT NULL DEFAULT '0',
  `luk` smallint(6) NOT NULL DEFAULT '0',
  `hp` smallint(6) NOT NULL DEFAULT '0',
  `mp` smallint(6) NOT NULL DEFAULT '0',
  `watk` smallint(6) NOT NULL DEFAULT '0',
  `matk` smallint(6) NOT NULL DEFAULT '0',
  `wdef` smallint(6) NOT NULL DEFAULT '0',
  `mdef` smallint(6) NOT NULL DEFAULT '0',
  `accuracy` smallint(6) NOT NULL DEFAULT '0',
  `avoid` smallint(6) NOT NULL DEFAULT '0',
  `hands` smallint(6) NOT NULL DEFAULT '0',
  `speed` smallint(6) NOT NULL DEFAULT '0',
  `jump` smallint(6) NOT NULL DEFAULT '0',
  `expireTime` bigint(20) NOT NULL DEFAULT '0',
  `creatorName` tinytext NOT NULL,
  PRIMARY KEY (`id`),
  KEY `characterID` (`characterID`),
  CONSTRAINT `items_ibfk_5` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


DROP TABLE IF EXISTS `skills`;
CREATE TABLE `skills` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `characterID` int(11) NOT NULL,
  `skillID` int(11) NOT NULL DEFAULT '0',
  `level` tinyint(4) NOT NULL DEFAULT '1',
  `cooldown` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_index` (`characterID`,`skillID`),
  KEY `characterID` (`characterID`),
  CONSTRAINT `skills_ibfk_2` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

DROP TABLE IF EXISTS `character_quests`;
CREATE TABLE `character_quests` (
  `characterID` INT(11) NOT NULL,
  `questID` SMALLINT(6) NOT NULL,
  `record` VARCHAR(255) NOT NULL DEFAULT '',
  `completed` TINYINT(1) NOT NULL DEFAULT '0',
  `completedAt` BIGINT(20) NOT NULL DEFAULT '0',
  PRIMARY KEY (`characterID`, `questID`),
  KEY `idx_character_quests_character` (`characterID`),
  KEY `idx_character_quests_completed` (`completed`),
  CONSTRAINT `character_quests_fk_character` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

DROP TABLE IF EXISTS `character_quest_kills`;
CREATE TABLE `character_quest_kills` (
  `characterID` INT NOT NULL,
  `questID` SMALLINT NOT NULL,
  `mobID` INT NOT NULL,
  `kills` INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`characterID`, `questID`, `mobID`), CONSTRAINT `c_q_kills_fk_character` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

DROP TABLE IF EXISTS `fame_log`;
CREATE TABLE `fame_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `from` int(11) NOT NULL,
  `to`   int(11) NOT NULL,
  `time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_from_time` (`from`, `time`),
  KEY `idx_to_time` (`to`, `time`),
  CONSTRAINT `fame_log_ibfk_from` FOREIGN KEY (`from`) REFERENCES `characters` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fame_log_ibfk_to`   FOREIGN KEY (`to`)   REFERENCES `characters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS account_storage (
    accountID   INT(10) UNSIGNED NOT NULL,
    slots       TINYINT UNSIGNED NOT NULL DEFAULT 20,
    mesos       INT(11) UNSIGNED NOT NULL DEFAULT 0,
    updatedAt   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (accountID),
    CONSTRAINT fk_storage_account
    FOREIGN KEY (accountID) REFERENCES accounts(accountID)
    ON DELETE CASCADE ON UPDATE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS account_storage_items (
    id           INT(11) NOT NULL AUTO_INCREMENT,
    accountID    INT(10) UNSIGNED NOT NULL,
    itemID       INT(11) NOT NULL,
    inventoryID  TINYINT(3) UNSIGNED NOT NULL,
    slotNumber   INT(11) NOT NULL,
    amount       INT(11) NOT NULL DEFAULT 1,
    flag         TINYINT(4) NOT NULL DEFAULT 0,
    upgradeSlots TINYINT(4) NOT NULL DEFAULT 0,
    level        TINYINT(4) NOT NULL DEFAULT 0,
    str          SMALLINT(6) NOT NULL DEFAULT 0,
    dex          SMALLINT(6) NOT NULL DEFAULT 0,
    intt         SMALLINT(6) NOT NULL DEFAULT 0,
    luk          SMALLINT(6) NOT NULL DEFAULT 0,
    hp           SMALLINT(6) NOT NULL DEFAULT 0,
    mp           SMALLINT(6) NOT NULL DEFAULT 0,
    watk         SMALLINT(6) NOT NULL DEFAULT 0,
    matk         SMALLINT(6) NOT NULL DEFAULT 0,
    wdef         SMALLINT(6) NOT NULL DEFAULT 0,
    mdef         SMALLINT(6) NOT NULL DEFAULT 0,
    accuracy     SMALLINT(6) NOT NULL DEFAULT 0,
    avoid        SMALLINT(6) NOT NULL DEFAULT 0,
    hands        SMALLINT(6) NOT NULL DEFAULT 0,
    speed        SMALLINT(6) NOT NULL DEFAULT 0,
    jump         SMALLINT(6) NOT NULL DEFAULT 0,
    expireTime   BIGINT(20) NOT NULL DEFAULT 0,
    creatorName  TINYTEXT NULL,
    PRIMARY KEY (id),
    KEY idx_storage_account (accountID),
    KEY idx_storage_tab_slot (accountID, inventoryID, slotNumber),
    CONSTRAINT fk_storage_items_account
    FOREIGN KEY (accountID) REFERENCES accounts(accountID)
    ON DELETE CASCADE ON UPDATE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS  `pets` (
    `parentID` INT(11) NOT NULL,
    `name` VARCHAR(64) NOT NULL,
    `sn` INT(11) NOT NULL,
    `level` TINYINT(3) UNSIGNED NOT NULL DEFAULT 1,
    `closeness` SMALLINT(6) UNSIGNED NOT NULL DEFAULT 0,
    `fullness` TINYINT(3) UNSIGNED NOT NULL DEFAULT 100,
    `deadDate` BIGINT(20) NOT NULL DEFAULT 0,
    `spawnDate` BIGINT(20) NOT NULL DEFAULT 0,
    `lastInteraction` BIGINT(20) NOT NULL DEFAULT 0,
    `spawned` BOOLEAN NOT NULL DEFAULT FALSE,
    `createdAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updatedAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`parentID`),
    CONSTRAINT `fk_pet_item` FOREIGN KEY (`parentID`)
        REFERENCES `items` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


-- 2025-08-19 16:51:40 UTC
