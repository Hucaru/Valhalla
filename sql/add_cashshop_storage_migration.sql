-- Migration to add cash shop storage tables
-- This storage is account-wide, similar to account_storage but for cash shop items

CREATE TABLE IF NOT EXISTS account_cashshop_storage (
    accountID   INT(10) UNSIGNED NOT NULL,
    slots       TINYINT UNSIGNED NOT NULL DEFAULT 50,
    updatedAt   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (accountID),
    CONSTRAINT fk_cashshop_storage_account
    FOREIGN KEY (accountID) REFERENCES accounts(accountID)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS account_cashshop_storage_items (
    id           BIGINT(20) NOT NULL AUTO_INCREMENT,
    accountID    INT(10) UNSIGNED NOT NULL,
    itemID       INT(11) NOT NULL,
    cashID       BIGINT(20) DEFAULT NULL,
    sn           INT(11) NOT NULL DEFAULT 0,
    slotNumber   INT(11) NOT NULL,
    amount       INT(11) NOT NULL DEFAULT 1,
    flag         TINYINT(4) NOT NULL DEFAULT 0,
    upgradeSlots TINYINT(4) NOT NULL DEFAULT 0,
    level        TINYINT(4) NOT NULL DEFAULT 0,
    str          SMALLINT(6) NOT NULL DEFAULT 0,
    dex          SMALLINT(6) NOT NULL DEFAULT 0,
    intt         SMALLINT(6) NOT NULL DEFAULT 0,
    luk          SMALLINT(6) NOT NULL DEFAULT 0,
    hp           SMALLINT(6) NOT NULL DEFAULT 0,
    mp           SMALLINT(6) NOT NULL DEFAULT 0,
    watk         SMALLINT(6) NOT NULL DEFAULT 0,
    matk         SMALLINT(6) NOT NULL DEFAULT 0,
    wdef         SMALLINT(6) NOT NULL DEFAULT 0,
    mdef         SMALLINT(6) NOT NULL DEFAULT 0,
    accuracy     SMALLINT(6) NOT NULL DEFAULT 0,
    avoid        SMALLINT(6) NOT NULL DEFAULT 0,
    hands        SMALLINT(6) NOT NULL DEFAULT 0,
    speed        SMALLINT(6) NOT NULL DEFAULT 0,
    jump         SMALLINT(6) NOT NULL DEFAULT 0,
    expireTime   BIGINT(20) NOT NULL DEFAULT 0,
    creatorName  TINYTEXT NULL,
    purchaseDate TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_cashshop_storage_account (accountID),
    KEY idx_cashshop_storage_slot (accountID, slotNumber),
    CONSTRAINT fk_cashshop_storage_items_account
    FOREIGN KEY (accountID) REFERENCES accounts(accountID)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

ALTER TABLE items
    ADD COLUMN cashID BIGINT(20) DEFAULT NULL AFTER creatorName,
ADD COLUMN cashSN INT(11) DEFAULT NULL AFTER cashID;
