package channel

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Hucaru/Valhalla/common"
)

type storageResultOp byte

const (
	storageInvFullOrNot   storageResultOp = 9
	storageNotEnoughMesos storageResultOp = 12
	storageIsFull         storageResultOp = 13
	storageDueToAnError   storageResultOp = 14
	storageSuccess        storageResultOp = 15
)

// Storage implements account-wide storage (no worldID scoping).
// It keeps an array of items per inventory tab (1..5) with MaxSlots capacity per tab.
type Storage struct {
	AccountID int32

	MaxSlots       byte
	TotalSlotsUsed byte
	Mesos          int32

	items map[byte][]Item // key = invID (1..5), value = slots [0..MaxSlots-1]
}

// NewStorage creates an empty storage instance (call Load before use).
func NewStorage(accountID int32) *Storage {
	return &Storage{
		AccountID: accountID,
		MaxSlots:  4,
		items:     make(map[byte][]Item, 5),
	}
}

// Load loads storage meta and items for the account.
// If storage row doesn't exist, it initializes a default one.
func (s *Storage) Load() error {
	if s.AccountID == 0 {
		return fmt.Errorf("storage.Load: invalid accountID")
	}

	// Ensure storage meta row exists
	var (
		slots sql.NullInt64
		mesos sql.NullInt64
	)
	err := common.DB.QueryRow(
		"SELECT slots, mesos FROM account_storage WHERE accountID=?",
		s.AccountID,
	).Scan(&slots, &mesos)
	switch {
	case err == sql.ErrNoRows:
		// Initialize default
		if _, ierr := common.DB.Exec(
			"INSERT INTO account_storage(accountID, slots, mesos) VALUES(?,?,?)",
			s.AccountID, 4, 0,
		); ierr != nil {
			return fmt.Errorf("storage.Load: init insert failed: %w", ierr)
		}
		s.MaxSlots = 4
		s.Mesos = 0
	case err != nil:
		return fmt.Errorf("storage.Load: select failed: %w", err)
	default:
		if slots.Valid {
			s.MaxSlots = clampByte(byte(slots.Int64), 4, 100)
		} else {
			s.MaxSlots = 4
		}
		if mesos.Valid {
			s.Mesos = int32(mesos.Int64)
		}
	}

	// Prepare slot arrays per tab
	if s.items == nil {
		s.items = make(map[byte][]Item, 5)
	}
	for inv := byte(1); inv <= 5; inv++ {
		s.items[inv] = make([]Item, s.MaxSlots)
	}
	s.TotalSlotsUsed = 0

	// Load items
	rows, qerr := common.DB.Query(`
		SELECT 
			id, itemID, inventoryID, slotNumber, amount,
			flag, upgradeSlots, level, str, dex, intt, luk, hp, mp,
			watk, matk, wdef, mdef, accuracy, avoid, hands, speed, jump,
			expireTime, creatorName
		FROM account_storage_items
		WHERE accountID=?
		ORDER BY inventoryID, slotNumber ASC`, s.AccountID)
	if qerr != nil {
		return fmt.Errorf("storage.Load: items query failed: %w", qerr)
	}
	defer rows.Close()

	for rows.Next() {
		var it Item
		var creator sql.NullString
		if err := rows.Scan(
			&it.dbID, &it.ID, &it.invID, &it.slotID, &it.amount,
			&it.flag, &it.upgradeSlots, &it.scrollLevel, &it.str, &it.dex, &it.intt, &it.luk, &it.hp, &it.mp,
			&it.watk, &it.matk, &it.wdef, &it.mdef, &it.accuracy, &it.avoid, &it.hands, &it.speed, &it.jump,
			&it.expireTime, &creator,
		); err != nil {
			log.Printf("storage.Load: scan item failed: %v", err)
			continue
		}
		if creator.Valid {
			it.creatorName = creator.String
		} else {
			it.creatorName = ""
		}

		// slotID is 1-based in inventory semantics; store at index (slot-1)
		if it.invID < 1 || it.invID > 5 {
			continue
		}
		if it.slotID <= 0 || byte(it.slotID) > s.MaxSlots {
			continue
		}
		idx := it.slotID - 1

		// Place into tab
		tab := s.items[it.invID]
		tab[idx] = it
		s.items[it.invID] = tab
		if it.dbID != 0 {
			s.TotalSlotsUsed++
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("storage.Load: rows error: %w", err)
	}

	return nil
}

// Save persists storage meta and items for the account.
// Simpler, robust strategy: rewrite all items for this account in a single transaction.
func (s *Storage) Save() error {
	if s.AccountID == 0 {
		return fmt.Errorf("storage.Save: invalid accountID")
	}

	tx, err := common.DB.Begin()
	if err != nil {
		return fmt.Errorf("storage.Save: begin tx failed: %w", err)
	}
	defer func() {
		_ = rollbackIfNeeded(tx, &err)
	}()

	// Update meta
	if _, uerr := tx.Exec(
		"UPDATE account_storage SET slots=?, mesos=? WHERE accountID=?",
		s.MaxSlots, s.Mesos, s.AccountID,
	); uerr != nil {
		err = fmt.Errorf("storage.Save: update meta failed: %w", uerr)
		return err
	}

	// Rewrite items (delete+bulk insert)
	if _, derr := tx.Exec(
		"DELETE FROM account_storage_items WHERE accountID=?",
		s.AccountID,
	); derr != nil {
		err = fmt.Errorf("storage.Save: delete items failed: %w", derr)
		return err
	}

	ins := `
		INSERT INTO account_storage_items(
			accountID, itemID, inventoryID, slotNumber, amount, flag, upgradeSlots, level,
			str, dex, intt, luk, hp, mp, watk, matk, wdef, mdef, accuracy, avoid, hands,
			speed, jump, expireTime, creatorName
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`
	stmt, perr := tx.Prepare(ins)
	if perr != nil {
		err = fmt.Errorf("storage.Save: prepare insert failed: %w", perr)
		return err
	}
	defer stmt.Close()

	// Ensure stable order
	type rowKey struct {
		inv  byte
		slot int16
	}
	keys := make([]rowKey, 0, int(s.MaxSlots*5))
	for inv := byte(1); inv <= 5; inv++ {
		for i := int16(1); i <= int16(s.MaxSlots); i++ {
			keys = append(keys, rowKey{inv: inv, slot: i})
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].inv != keys[j].inv {
			return keys[i].inv < keys[j].inv
		}
		return keys[i].slot < keys[j].slot
	})

	for _, k := range keys {
		idx := k.slot - 1
		tab := s.items[k.inv]
		if tab == nil || int(idx) >= len(tab) {
			continue
		}
		it := tab[idx]
		if it.ID == 0 || it.amount == 0 {
			continue
		}
		if _, ierr := stmt.Exec(
			s.AccountID, it.ID, it.invID, it.slotID, it.amount, it.flag, it.upgradeSlots, it.scrollLevel,
			it.str, it.dex, it.intt, it.luk, it.hp, it.mp, it.watk, it.matk, it.wdef, it.mdef, it.accuracy, it.avoid, it.hands,
			it.speed, it.jump, it.expireTime, nullableStr(it.creatorName),
		); ierr != nil {
			err = fmt.Errorf("storage.Save: insert item (%d) failed: %w", it.ID, ierr)
			return err
		}
	}

	if cerr := tx.Commit(); cerr != nil {
		err = fmt.Errorf("storage.Save: commit failed: %w", cerr)
		return err
	}
	return nil
}

// AddItem inserts an item into the first empty slot of its inventory tab.
// Returns true if successful.
func (s *Storage) AddItem(it Item) bool {
	inv := it.invID
	if inv < 1 || inv > 5 {
		return false
	}
	tab := s.items[inv]
	for i := 0; i < int(s.MaxSlots); i++ {
		if tab[i].ID != 0 {
			continue
		}
		s.TotalSlotsUsed++
		it.slotID = int16(i + 1)
		tab[i] = it
		s.items[inv] = tab
		return true
	}
	return false
}

// GetInventoryItems returns a copy of all non-empty items in the tab.
func (s *Storage) GetInventoryItems(inv byte) []Item {
	out := make([]Item, 0, s.MaxSlots)
	tab := s.items[inv]
	for i := range tab {
		if tab[i].ID != 0 {
			out = append(out, tab[i])
		}
	}
	return out
}

// TakeItemOut removes the item at a slot (0-based) in the given tab,
// then compacts the tab array so clients see contiguous items.
func (s *Storage) TakeItemOut(inv byte, slotZeroBased byte) {
	if inv < 1 || inv > 5 {
		return
	}
	if slotZeroBased >= s.MaxSlots {
		return
	}
	tab := s.items[inv]
	newTab := make([]Item, s.MaxSlots)
	dst := 0
	for i := 0; i < int(s.MaxSlots); i++ {
		if i == int(slotZeroBased) {
			if tab[i].ID != 0 && s.TotalSlotsUsed > 0 {
				s.TotalSlotsUsed--
			}
			continue
		}
		if tab[i].ID != 0 {
			item := tab[i]
			dst++
			item.slotID = int16(dst)
			newTab[dst-1] = item
		}
	}
	s.items[inv] = newTab
}

// GetItem returns the item pointer at a given tab and 0-based slot, or nil if empty/out of range.
func (s *Storage) GetItem(inv byte, slotZeroBased byte) *Item {
	if inv < 1 || inv > 5 {
		return nil
	}
	if slotZeroBased >= s.MaxSlots {
		return nil
	}
	tab := s.items[inv]
	if int(slotZeroBased) >= len(tab) {
		return nil
	}
	if tab[slotZeroBased].ID == 0 {
		return nil
	}
	return &tab[slotZeroBased]
}

// SetSlots resizes per-tab capacity. Min 4, Max 100. Fails if requested below current usage.
func (s *Storage) SetSlots(amount byte) bool {
	amount = clampByte(amount, 4, 100)
	if amount < s.TotalSlotsUsed {
		return false
	}
	s.MaxSlots = amount
	for inv := byte(1); inv <= 5; inv++ {
		tab := s.items[inv]
		if tab == nil {
			s.items[inv] = make([]Item, s.MaxSlots)
			continue
		}
		if byte(len(tab)) == s.MaxSlots {
			continue
		}
		newTab := make([]Item, s.MaxSlots)
		copy(newTab, tab)
		s.items[inv] = newTab
	}
	return true
}

// SlotsAvailable returns true if there is at least one empty slot across all tabs.
func (s *Storage) SlotsAvailable() bool {
	return s.MaxSlots > s.TotalSlotsUsed
}

// ChangeMesos adjusts mesos by delta. Clamps to [0, int32_max].
func (s *Storage) ChangeMesos(delta int32) {
	newVal := int64(s.Mesos) + int64(delta)
	if newVal < 0 {
		newVal = 0
	} else if newVal > int64(^uint32(0)>>1) { // int32 max
		newVal = int64(^uint32(0) >> 1)
	}
	s.Mesos = int32(newVal)
}

// Helpers

func clampByte(v, min, max byte) byte {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func rollbackIfNeeded(tx *sql.Tx, perr *error) error {
	if *perr == nil {
		return nil
	}
	_ = tx.Rollback()
	return nil
}

func nullableStr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
