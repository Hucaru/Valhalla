-- Adminer 4.3.1 MySQL dump

SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `userID` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `username` text NOT NULL,
  `password` text NOT NULL,
  `isLogedIn` int(11) NOT NULL DEFAULT '0',
  `isAdmin` int(11) NOT NULL DEFAULT '0',
  `isBanned` int(11) NOT NULL,
  PRIMARY KEY (`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

INSERT INTO `users` (`userID`, `username`, `password`, `isLogedIn`, `isAdmin`, `isBanned`) VALUES
(1,	'test',	'125d6d03b32c84d492747f79cf0bf6e179d287f341384eb5d6d3197525ad6be8e6df0116032935698f99a09e265073d1d6c32c274591bf1d0a20ad67cba921bc',	1,	1,	0);

-- 2017-08-25 23:50:39
