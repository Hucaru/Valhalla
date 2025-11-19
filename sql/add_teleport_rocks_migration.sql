-- Migration script to add teleport rocks columns to characters table
-- This script is for upgrading existing databases to support teleport rocks feature

-- Drop old column if it exists (from previous version)
ALTER TABLE `characters` DROP COLUMN IF EXISTS `teleportRocks`;

-- Add the regTeleportRocks column if it doesn't exist (5 slots)
ALTER TABLE `characters` 
ADD COLUMN IF NOT EXISTS `regTeleportRocks` text NOT NULL DEFAULT '';

-- Add the vipTeleportRocks column if it doesn't exist (10 slots)
ALTER TABLE `characters` 
ADD COLUMN IF NOT EXISTS `vipTeleportRocks` text NOT NULL DEFAULT '';

-- Initialize empty regular teleport rocks for all existing characters (5 slots)
UPDATE `characters` 
SET `regTeleportRocks` = '999999999,999999999,999999999,999999999,999999999'
WHERE `regTeleportRocks` = '' OR `regTeleportRocks` IS NULL;

-- Initialize empty VIP teleport rocks for all existing characters (10 slots)
UPDATE `characters` 
SET `vipTeleportRocks` = '999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999'
WHERE `vipTeleportRocks` = '' OR `vipTeleportRocks` IS NULL;
