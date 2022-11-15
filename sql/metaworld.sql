SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `accounts`;
CREATE TABLE `accounts` (
  `accountID` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `uId` tinytext NOT NULL,
  `username` tinytext NOT NULL,
  `password` tinytext NOT NULL,
  `pin` tinytext NOT NULL,
  `isLogedIn` tinyint(4) NOT NULL DEFAULT '0',
  `adminLevel` tinyint(4) NOT NULL DEFAULT '0',
  `isBanned` int(11) NOT NULL DEFAULT '0',
  `gender` tinyint(4) NOT NULL DEFAULT '0',
  `dob` int(11) NOT NULL,
  PRIMARY KEY (`accountID`),
  UNIQUE KEY `unique_index_uID` (`uId`)
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
  CONSTRAINT `buddy_ibfk_1` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`),
  CONSTRAINT `buddy_ibfk_2` FOREIGN KEY (`friendID`) REFERENCES `characters` (`id`) ON DELETE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


DROP TABLE IF EXISTS `characters`;
CREATE TABLE `characters` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `accountID` int(10) unsigned NOT NULL,
  `worldID` int(11) unsigned NOT NULL DEFAULT '1',
  `channelID` tinyint(2) NOT NULL DEFAULT '1',
  `migrationID` tinyint(4) NOT NULL DEFAULT '-1',
  `nickname` tinytext NOT NULL,
  `gender` int(11) unsigned NOT NULL DEFAULT '1',
  `hair` tinytext ,
  `top` tinytext,
  `bottom` tinytext,
  `clothes` tinytext,
  `role` int(11) unsigned NOT NULL DEFAULT '0',
  `mapID` int(11) unsigned NOT NULL DEFAULT '0',
  `previousMapID` int(11) unsigned NOT NULL DEFAULT '0',
  `createdAt` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_index_nickname` (`nickname`),
  KEY `userID` (`accountID`),
  CONSTRAINT `characters_ibfk_1` FOREIGN KEY (`accountID`) REFERENCES `accounts` (`accountID`)
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


DROP TABLE IF EXISTS `movement`;
CREATE TABLE `movement` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `characterID` int(11) NOT NULL,
  `pos_x` float NOT NULL DEFAULT '0',
  `pos_y` float NOT NULL DEFAULT '0',
  `pos_z` float NOT NULL DEFAULT '0',
  `rot_x` float NOT NULL DEFAULT '0',
  `rot_y` float NOT NULL DEFAULT '0',
  `rot_z` float NOT NULL DEFAULT '0',
  `time` bigint(20) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `characterID` (`characterID`),
  CONSTRAINT `movement_ibfk_2` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


DROP TABLE IF EXISTS `chat`;
CREATE TABLE `chat` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `characterID` int(11) NOT NULL,
    `regionID` int(11) NOT NULL DEFAULT '-1',
    `text` text NOT NULL DEFAULT '',
    `targetID` int(11) NOT NULL,
    `createdAt` bigint(20) NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    KEY `characterID` (`characterID`),
    CONSTRAINT `chat_ibfk_2` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;