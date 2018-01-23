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
  `isMigratingWorld` tinyint(4) NOT NULL DEFAULT '-1',
  `isMigratingChannel` tinyint(4) NOT NULL DEFAULT '-1',
  `name` text NOT NULL,
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
  `equipSlotSize` tinyint(4) NOT NULL DEFAULT '50',
  `useSlotSize` tinyint(4) NOT NULL DEFAULT '50',
  `setupSlotSize` tinyint(4) NOT NULL DEFAULT '50',
  `etcSlotSize` tinyint(4) NOT NULL DEFAULT '50',
  `cashSlotSize` tinyint(4) NOT NULL DEFAULT '50',
  PRIMARY KEY (`id`),
  KEY `userID` (`userID`),
  CONSTRAINT `characters_ibfk_1` FOREIGN KEY (`userID`) REFERENCES `users` (`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

INSERT INTO `characters` (`id`, `userID`, `worldID`, `isMigratingWorld`, `isMigratingChannel`, `name`, `gender`, `skin`, `hair`, `face`, `level`, `job`, `str`, `dex`, `intt`, `luk`, `hp`, `maxHP`, `mp`, `maxMP`, `ap`, `sp`, `exp`, `fame`, `mapID`, `mapPos`, `previousMapID`, `mesos`, `equipSlotSize`, `useSlotSize`, `setupSlotSize`, `etcSlotSize`, `cashSlotSize`) VALUES
(8,	1,	0,	0,	1,	'[GM]Hucaru',	0,	0,	30020,	20000,	200,	510,	7,	5,	6,	7,	100,	100,	50,	50,	0,	0,	0,	1001,	100000000,	0,	0,	100,	50,	50,	50,	50,	50),
(15,	2,	0,	0,	0,	'Hucaru',	0,	0,	30030,	20000,	1,	0,	5,	5,	10,	5,	100,	100,	50,	50,	0,	0,	0,	0,	0,	0,	0,	0,	50,	50,	50,	50,	50),
(17,	1,	1,	-1,	-1,	'[GM]Bera',	0,	0,	30023,	20000,	1,	0,	8,	5,	6,	6,	100,	100,	50,	50,	0,	0,	0,	0,	0,	0,	0,	0,	50,	50,	50,	50,	50);

DROP TABLE IF EXISTS `equips`;
CREATE TABLE `equips` (
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
  CONSTRAINT `equips_ibfk_3` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

INSERT INTO `equips` (`id`, `characterID`, `itemID`, `slotNumber`, `upgradeSlots`, `level`, `str`, `dex`, `intt`, `luk`, `hp`, `mp`, `watk`, `matk`, `wdef`, `mdef`, `accuracy`, `avoid`, `hands`, `speed`, `jump`, `expireTime`) VALUES
(30,	8,	1002140,	-1,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0),
(31,	8,	1042003,	-5,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0),
(32,	8,	1062007,	-6,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0),
(33,	8,	1072004,	-7,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0),
(34,	8,	1322013,	-11,	0,	0,	0,	0,	0,	0,	0,	0,	200,	200,	200,	200,	200,	200,	200,	0,	0,	0),
(35,	8,	1082002,	-8,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0),
(36,	8,	1102054,	-9,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0),
(37,	8,	1092008,	-10,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0),
(41,	8,	1002342,	-101,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0),
(45,	8,	1042001,	-105,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	134567899),
(52,	8,	1112101,	-116,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0),
(53,	8,	1702027,	-111,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	150842304000000000),
(54,	8,	1002140,	1,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	150842304000000000);

DROP TABLE IF EXISTS `items`;
CREATE TABLE `items` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `characterID` int(11) NOT NULL,
  `inventoryID` tinyint(4) NOT NULL DEFAULT '1',
  `itemID` int(11) NOT NULL,
  `slotNumber` int(11) NOT NULL,
  `upgradeSlots` tinyint(4) NOT NULL DEFAULT '0',
  `level` tinyint(4) NOT NULL DEFAULT '0',
  `amount` int(11) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `characterID` (`characterID`),
  CONSTRAINT `items_ibfk_1` FOREIGN KEY (`characterID`) REFERENCES `characters` (`id`)
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
  `isInChannel` tinyint(4) NOT NULL DEFAULT '-1',
  PRIMARY KEY (`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

INSERT INTO `users` (`userID`, `username`, `password`, `isLogedIn`, `isAdmin`, `isBanned`, `gender`, `dob`, `isInChannel`) VALUES
(1,	'test',	'125d6d03b32c84d492747f79cf0bf6e179d287f341384eb5d6d3197525ad6be8e6df0116032935698f99a09e265073d1d6c32c274591bf1d0a20ad67cba921bc',	0,	1,	0,	0,	19900101,	0),
(2,	'test2',	'125d6d03b32c84d492747f79cf0bf6e179d287f341384eb5d6d3197525ad6be8e6df0116032935698f99a09e265073d1d6c32c274591bf1d0a20ad67cba921bc',	0,	0,	0,	0,	19900101,	0);

-- 2018-01-23 23:35:01
