package cashshop

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/common"
)

// Cash shop storage capacity bounds
const (
	cashShopStorageMinSlots byte = 50
	cashShopStorageMaxSlots byte = 255
)

// CashShopStorage represents account-wide cash shop storage
type CashShopStorage struct {
	accountID      int32
	maxSlots       byte
	totalSlotsUsed byte
	items          []channel.Item
}

func clampByte(v, min, max byte) byte {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// NewCashShopStorage creates a new cash shop storage instance
func NewCashShopStorage(accountID int32) *CashShopStorage {
	return &CashShopStorage{
		accountID: accountID,
		maxSlots:  cashShopStorageMinSlots,
		items:     make([]channel.Item, cashShopStorageMinSlots),
	}
}

func (s *CashShopStorage) ensureCapacity() {
	if s.items == nil || byte(len(s.items)) != s.maxSlots {
		newArr := make([]channel.Item, s.maxSlots)
		if s.items != nil {
			copy(newArr, s.items)
		}
		s.items = newArr
	}
}

// Load items and header from DB into channel.Item
func (s *CashShopStorage) load() error {
	var slots sql.NullInt64
	if err := common.DB.QueryRow(
		"SELECT slots FROM account_cashshop_storage WHERE accountID=?",
		s.accountID,
	).Scan(&slots); err != nil {
		if err == sql.ErrNoRows {
			if _, ierr := common.DB.Exec(
				"INSERT INTO account_cashshop_storage(accountID, slots) VALUES(?,?)",
				s.accountID, cashShopStorageMinSlots,
			); ierr != nil {
				return fmt.Errorf("couldn't initialize cash shop storage for account %d: %w", s.accountID, ierr)
			}
			s.maxSlots = cashShopStorageMinSlots
			s.ensureCapacity()
			s.totalSlotsUsed = 0
			return nil
		}
		return fmt.Errorf("failed to load cash shop storage header for account %d: %w", s.accountID, err)
	}

	if slots.Valid {
		s.maxSlots = clampByte(byte(slots.Int64), cashShopStorageMinSlots, cashShopStorageMaxSlots)
	} else {
		s.maxSlots = cashShopStorageMinSlots
	}

	s.ensureCapacity()
	s.totalSlotsUsed = 0

	rows, qerr := common.DB.Query(`
		SELECT 
			itemID, cashID, sn, slotNumber, amount,
			flag, upgradeSlots, level, str, dex, intt, luk, hp, mp,
			watk, matk, wdef, mdef, accuracy, avoid, hands, speed, jump,
			expireTime, creatorName, UNIX_TIMESTAMP(purchaseDate)
		FROM account_cashshop_storage_items
		WHERE accountID=?
		ORDER BY slotNumber ASC`, s.accountID)
	if qerr != nil {
		return fmt.Errorf("failed to load cash shop storage items for account %d: %w", s.accountID, qerr)
	}
	defer rows.Close()

	for rows.Next() {
		var slotNumber int16
		var itemID int32
		var amount int16
		var flag int16
		var upgradeSlots byte
		var scrollLevel byte
		var str, dex, intt, luk, hp, mp int16
		var watk, matk, wdef, mdef int16
		var accuracy, avoid, hands, speed, jump int16
		var expireTime int64
		var creatorName sql.NullString
		var cashIDNullable sql.NullInt64
		var sn int32
		var _purchaseTS sql.NullInt64

		if err := rows.Scan(
			&itemID, &cashIDNullable, &sn, &slotNumber, &amount,
			&flag, &upgradeSlots, &scrollLevel,
			&str, &dex, &intt, &luk,
			&hp, &mp, &watk, &matk,
			&wdef, &mdef, &accuracy, &avoid,
			&hands, &speed, &jump,
			&expireTime, &creatorName, &_purchaseTS,
		); err != nil {
			log.Println("Error scanning cash shop storage item:", err)
			continue
		}

		var creator string
		if creatorName.Valid {
			creator = creatorName.String
		}

		it, ierr := channel.CreateItemFromDBValues(
			itemID, slotNumber, amount, flag, upgradeSlots, scrollLevel,
			str, dex, intt, luk, hp, mp, watk, matk, wdef, mdef,
			accuracy, avoid, hands, speed, jump, expireTime, creator,
		)
		if ierr != nil {
			log.Println("Error creating item from DB values:", ierr)
			continue
		}

		if cashIDNullable.Valid {
			it.SetCashID(cashIDNullable.Int64)
		} else {
			it.SetCashID(channel.GenerateCashID())
		}
		it.SetCashSN(sn)

		if slotNumber <= 0 || slotNumber > int16(s.maxSlots) {
			continue
		}
		idx := int(slotNumber - 1)
		s.items[idx] = it
		if itemID != 0 {
			s.totalSlotsUsed++
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error while reading cash shop storage items for account %d: %w", s.accountID, err)
	}

	return nil
}

func (s *CashShopStorage) save() (err error) {
	tx, terr := common.DB.Begin()
	if terr != nil {
		return fmt.Errorf("couldn't open transaction to save cash shop storage (acct %d): %w", s.accountID, terr)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, uerr := tx.Exec(
		"UPDATE account_cashshop_storage SET slots=? WHERE accountID=?",
		s.maxSlots, s.accountID,
	); uerr != nil {
		err = fmt.Errorf("failed to update cash shop storage header (acct %d): %w", s.accountID, uerr)
		return
	}

	if _, derr := tx.Exec(
		"DELETE FROM account_cashshop_storage_items WHERE accountID=?",
		s.accountID,
	); derr != nil {
		err = fmt.Errorf("failed to clear cash shop storage items (acct %d): %w", s.accountID, derr)
		return
	}

	for i := range s.items {
		if s.items[i].ID == 0 {
			continue
		}

		slotNumber := int16(i + 1)
		if ierr := s.items[i].SaveToCashShopStorage(tx, s.accountID, slotNumber); ierr != nil {
			err = fmt.Errorf("failed inserting cash shop item (acct %d, slot %d): %w", s.accountID, slotNumber, ierr)
			return
		}
	}

	if cerr := tx.Commit(); cerr != nil {
		err = fmt.Errorf("failed to commit cash shop storage save (acct %d): %w", s.accountID, cerr)
		return
	}

	return nil
}

// addItem adds an item with a generated cashID and provided SN
func (s *CashShopStorage) addItem(item channel.Item, sn int32) (int, bool) {
	for i := 0; i < int(s.maxSlots); i++ {
		if s.items[i].ID != 0 {
			continue
		}
		item.SetCashID(channel.GenerateCashID())
		item.SetCashSN(sn)
		s.items[i] = item
		s.totalSlotsUsed++
		return i, true
	}
	return -1, false
}

// addItemWithCashID adds an item with a specific cashID (used when returning from inventory)
func (s *CashShopStorage) addItemWithCashID(item channel.Item, sn int32, cashID int64) (int, bool) {
	for i := 0; i < int(s.maxSlots); i++ {
		if s.items[i].ID != 0 {
			continue
		}
		item.SetCashID(cashID)
		item.SetCashSN(sn)
		s.items[i] = item
		s.totalSlotsUsed++
		return i, true
	}
	return -1, false
}

// removeAt removes by index and returns the item
func (s *CashShopStorage) removeAt(idx int) (*channel.Item, bool) {
	if idx < 0 || idx >= int(s.maxSlots) {
		return nil, false
	}
	if s.items[idx].ID == 0 {
		return nil, false
	}

	item := s.items[idx]
	s.items[idx] = channel.Item{}
	if s.totalSlotsUsed > 0 {
		s.totalSlotsUsed--
	}
	return &item, true
}

// getAllItems returns non-empty items
func (s *CashShopStorage) getAllItems() []channel.Item {
	out := make([]channel.Item, 0, s.totalSlotsUsed)
	for i := range s.items {
		if s.items[i].ID != 0 {
			out = append(out, s.items[i])
		}
	}
	return out
}

func (s *CashShopStorage) slotsAvailable() bool {
	return s.totalSlotsUsed < s.maxSlots
}

func (s *CashShopStorage) getItemBySlot(slot int16) (*channel.Item, bool) {
	if slot <= 0 || int(slot) > len(s.items) {
		return nil, false
	}
	idx := int(slot - 1)
	if s.items[idx].ID == 0 {
		return nil, false
	}
	return &s.items[idx], true
}
