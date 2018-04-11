-- Adminer 4.6.2 MySQL dump

SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `characters`;
CREATE TABLE `characters` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userID` int(10) unsigned NOT NULL,
  `worldID` int(11) unsigned NOT NULL,
  `isMigratingWorld` tinyint(4) NOT NULL DEFAULT '-1',
  `isMigratingChannel` tinyint(4) NOT NULL DEFAULT '-1',
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
  `fame` int(11) unsigned NOT NULL DEFAULT '0',
  `mapID` int(11) unsigned NOT NULL DEFAULT '0',
  `mapPos` int(11) unsigned NOT NULL DEFAULT '0',
  `previousMapID` int(11) unsigned NOT NULL DEFAULT '0',
  `mesos` int(11) NOT NULL DEFAULT '0',
  `equipSlotSize` tinyint(4) NOT NULL DEFAULT '32',
  `useSlotSize` tinyint(4) NOT NULL DEFAULT '32',
  `setupSlotSize` tinyint(4) NOT NULL DEFAULT '32',
  `etcSlotSize` tinyint(4) NOT NULL DEFAULT '32',
  `cashSlotSize` tinyint(4) NOT NULL DEFAULT '32',
  PRIMARY KEY (`id`),
  KEY `userID` (`userID`),
  CONSTRAINT `characters_ibfk_1` FOREIGN KEY (`userID`) REFERENCES `users` (`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


DROP TABLE IF EXISTS `items`;
CREATE TABLE `items` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `characterID` int(11) NOT NULL,
  `itemID` int(11) NOT NULL,
  `inventoryID` int(11) NOT NULL DEFAULT '0',
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
  `level` tinyint(4) NOT NULL DEFAULT '0',
  `cooldown` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `characterID` (`characterID`),
  CONSTRAINT `skills_ibfk_2` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `userID` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` tinytext NOT NULL,
  `password` tinytext NOT NULL,
  `isLogedIn` tinyint(4) NOT NULL DEFAULT '0',
  `isAdmin` tinyint(4) NOT NULL DEFAULT '0',
  `isBanned` int(11) NOT NULL DEFAULT '0',
  `gender` tinyint(4) NOT NULL DEFAULT '0',
  `dob` int(11) NOT NULL,
  `isInChannel` tinyint(4) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


-- 2018-04-11 02:25:12
