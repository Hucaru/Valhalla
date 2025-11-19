-- Migration script to add teleportRocks column to characters table
-- This script is for upgrading existing databases to support teleport rocks feature

-- Add the teleportRocks column if it doesn't exist
ALTER TABLE `characters` 
ADD COLUMN IF NOT EXISTS `teleportRocks` text NOT NULL DEFAULT '';

-- Initialize empty teleport rocks for all existing characters
-- This ensures all characters have the teleportRocks field populated
UPDATE `characters` 
SET `teleportRocks` = '999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999,999999999'
WHERE `teleportRocks` = '' OR `teleportRocks` IS NULL;
