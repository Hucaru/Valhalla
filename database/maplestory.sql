-- Adminer 4.3.1 MySQL dump

SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `characters`;
CREATE TABLE `characters` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userID` int(10) unsigned NOT NULL,
  `worldID` int(11) unsigned NOT NULL,
  `name` text NOT NULL,
  `gender` int(11) unsigned NOT NULL,
  `skin` int(11) unsigned NOT NULL,
  `hair` int(11) unsigned NOT NULL,
  `face` int(11) unsigned NOT NULL,
  `level` int(200) unsigned NOT NULL DEFAULT '1',
  `job` int(11) unsigned NOT NULL DEFAULT '0',
  `str` int(11) unsigned NOT NULL,
  `dex` int(11) unsigned NOT NULL,
  `int` int(11) unsigned NOT NULL,
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
  `equipSlotSize` tinyint(4) NOT NULL DEFAULT '50',
  `useSlotSize` tinyint(4) NOT NULL DEFAULT '50',
  `setupSlotSize` tinyint(4) NOT NULL DEFAULT '50',
  `etcSlotSize` tinyint(4) NOT NULL DEFAULT '50',
  `cashSlotSize` tinyint(4) NOT NULL DEFAULT '50',
  PRIMARY KEY (`id`),
  KEY `userID` (`userID`),
  CONSTRAINT `characters_ibfk_1` FOREIGN KEY (`userID`) REFERENCES `users` (`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


DROP TABLE IF EXISTS `items`;
CREATE TABLE `items` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `characterID` int(11) NOT NULL,
  `itemID` int(11) NOT NULL,
  `slotNumber` int(11) NOT NULL,
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
  PRIMARY KEY (`id`),
  KEY `characterID` (`characterID`),
  CONSTRAINT `items_ibfk_3` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `userID` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` text NOT NULL,
  `password` text NOT NULL,
  `isLogedIn` tinyint(4) NOT NULL DEFAULT '0',
  `isAdmin` tinyint(4) NOT NULL DEFAULT '0',
  `isBanned` int(11) NOT NULL,
  `gender` tinyint(4) NOT NULL DEFAULT '0',
  `dob` int(11) NOT NULL,
  `isInChannel` tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

INSERT INTO `users` (`userID`, `username`, `password`, `isLogedIn`, `isAdmin`, `isBanned`, `gender`, `dob`, `isInChannel`) VALUES
(1,	'test',	'125d6d03b32c84d492747f79cf0bf6e179d287f341384eb5d6d3197525ad6be8e6df0116032935698f99a09e265073d1d6c32c274591bf1d0a20ad67cba921bc',	0,	1,	0,	0,	19900101,	0),
(2,	'test2',	'125d6d03b32c84d492747f79cf0bf6e179d287f341384eb5d6d3197525ad6be8e6df0116032935698f99a09e265073d1d6c32c274591bf1d0a20ad67cba921bc',	0,	0,	0,	0,	19900101,	0);

-- 2018-01-21 21:39:19
