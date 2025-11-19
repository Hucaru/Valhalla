-- Migration script to add teleport rocks columns to characters table
-- This script is for upgrading existing databases to support teleport rocks feature

-- Add the regTeleportRocks column (5 slots)
ALTER TABLE `characters` 
ADD COLUMN `regTeleportRocks` text NOT NULL;

-- Add the vipTeleportRocks column (10 slots)
ALTER TABLE `characters` 
ADD COLUMN `vipTeleportRocks` text NOT NULL;

-- Initialize empty regular teleport rocks for all existing characters (5 slots)
UPDATE `characters` 
SET `regTeleportRocks` = '999999999,999999999,999999999,999999999,999999999'
WHERE `regTeleportRocks` = '' OR `regTeleportRocks` IS NULL;

-- Initialize empty VIP teleport rocks for all existing characters (10 slots)
UPDATE `characters` 
SET `vipTeleportRocks` = '999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999'
WHERE `vipTeleportRocks` = '' OR `vipTeleportRocks` IS NULL;
