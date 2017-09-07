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
  `level` int(11) unsigned NOT NULL DEFAULT '1',
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
  PRIMARY KEY (`id`),
  KEY `userID` (`userID`),
  CONSTRAINT `characters_ibfk_1` FOREIGN KEY (`userID`) REFERENCES `users` (`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

INSERT INTO `characters` (`id`, `userID`, `worldID`, `name`, `gender`, `skin`, `hair`, `face`, `level`, `job`, `str`, `dex`, `int`, `luk`, `hp`, `maxHP`, `mp`, `maxMP`, `ap`, `sp`, `exp`, `fame`, `mapID`, `mapPos`, `previousMapID`) VALUES
(8,	1,	0,	'[GM]Hucaru',	0,	0,	30020,	20000,	1,	0,	7,	5,	6,	7,	100,	100,	50,	50,	0,	0,	0,	0,	0,	0,	0);

-- 2018-01-13 21:50:01
