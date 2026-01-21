-- Minimal Anti-Cheat: Just 2 tables, everything else in memory

-- Bans table (stores all ban records)
CREATE TABLE IF NOT EXISTS `bans` (
  `id` INT(11) NOT NULL AUTO_INCREMENT,
  `accountID` INT(10) UNSIGNED NULL,
  `reason` TEXT NOT NULL,
  `banEnd` TIMESTAMP NULL DEFAULT NULL COMMENT 'NULL = permanent',
  `ip` VARCHAR(45) DEFAULT NULL,
  `hwid` VARCHAR(20) DEFAULT NULL,
  `createdAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_account` (`accountID`, `banEnd`),
  KEY `idx_ip` (`ip`, `banEnd`),
  KEY `idx_hwid` (`hwid`, `banEnd`),
  CONSTRAINT `bans_fk_account` FOREIGN KEY (`accountID`) REFERENCES `accounts` (`accountID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Ban escalation tracking (temp ban count per account)
CREATE TABLE IF NOT EXISTS `ban_escalation` (
  `accountID` INT(10) UNSIGNED NOT NULL,
  `count` INT(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`accountID`),
  CONSTRAINT `ban_escalation_fk_account` FOREIGN KEY (`accountID`) REFERENCES `accounts` (`accountID`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Add HWID to accounts table (after maplepoints)
ALTER TABLE `accounts` ADD COLUMN `hwid` VARCHAR(20) DEFAULT NULL AFTER `maplepoints`;
ALTER TABLE `accounts` ADD COLUMN `isLocked`  int(11) NOT NULL DEFAULT '0' AFTER `isBanned`;
ALTER TABLE `accounts` ADD INDEX `idx_hwid` (`hwid`);