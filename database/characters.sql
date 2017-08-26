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
  `level` int(11) unsigned NOT NULL,
  `job` int(11) unsigned NOT NULL,
  `str` int(11) unsigned NOT NULL,
  `dex` int(11) unsigned NOT NULL,
  `int` int(11) unsigned NOT NULL,
  `luk` int(11) unsigned NOT NULL,
  `hp` int(11) unsigned NOT NULL,
  `maxHP` int(11) unsigned NOT NULL,
  `mp` int(11) unsigned NOT NULL,
  `maxMP` int(11) unsigned NOT NULL,
  `ap` int(11) unsigned NOT NULL,
  `sp` int(11) unsigned NOT NULL,
  `exp` int(11) unsigned NOT NULL,
  `fame` int(11) unsigned NOT NULL,
  `mapID` int(11) unsigned NOT NULL,
  `mapPos` int(11) unsigned NOT NULL,
  `previousMapID` int(11) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `userID` (`userID`),
  CONSTRAINT `characters_ibfk_1` FOREIGN KEY (`userID`) REFERENCES `users` (`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

INSERT INTO `characters` (`id`, `userID`, `worldID`, `name`, `gender`, `skin`, `hair`, `face`, `level`, `job`, `str`, `dex`, `int`, `luk`, `hp`, `maxHP`, `mp`, `maxMP`, `ap`, `sp`, `exp`, `fame`, `mapID`, `mapPos`, `previousMapID`) VALUES
(1,	1,	1,	'test',	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0,	0);

-- 2017-08-26 18:21:35
