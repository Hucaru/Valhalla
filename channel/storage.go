package channel

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/Hucaru/Valhalla/common"
)

// Storage capacity bounds
const (
	storageMinSlots byte = 20
	storageMaxSlots byte = 255
)

type storage struct {
	maxSlots       byte
	totalSlotsUsed byte
	mesos          int32

	items []Item
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

func (s *storage) ensureCapacity() {
	if s.items == nil || byte(len(s.items)) != s.maxSlots {
		newArr := make([]Item, s.maxSlots)
		if s.items != nil {
			copy(newArr, s.items)
		}
		s.items = newArr
	}
}

func (s *storage) load(accountID int32) error {
	var slots, mesos sql.NullInt64
	if err := common.DB.QueryRow(
		"SELECT slots, mesos FROM account_storage WHERE accountID=?",
		accountID,
	).Scan(&slots, &mesos); err != nil {
		if err == sql.ErrNoRows {
			if _, ierr := common.DB.Exec(
				"INSERT INTO account_storage(accountID, slots, mesos) VALUES(?,?,?)",
				accountID, storageMinSlots, 0,
			); ierr != nil {
				return fmt.Errorf("couldn't initialize storage for account %d: %w", accountID, ierr)
			}
			s.maxSlots = storageMinSlots
			s.mesos = 0
			s.ensureCapacity()
			s.totalSlotsUsed = 0
			return nil
		}
		return fmt.Errorf("failed to load storage header for account %d: %w", accountID, err)
	}

	if slots.Valid {
		s.maxSlots = clampByte(byte(slots.Int64), storageMinSlots, storageMaxSlots)
	} else {
		s.maxSlots = storageMinSlots
	}
	if mesos.Valid {
		s.mesos = int32(mesos.Int64)
	}

	s.ensureCapacity()
	s.totalSlotsUsed = 0

	rows, qerr := common.DB.Query(`
		SELECT 
			id, itemID, inventoryID, slotNumber, amount,
			flag, upgradeSlots, level, str, dex, intt, luk, hp, mp,
			watk, matk, wdef, mdef, accuracy, avoid, hands, speed, jump,
			expireTime, creatorName
		FROM account_storage_items
		WHERE accountID=?
		ORDER BY slotNumber ASC`, accountID)
	if qerr != nil {
		return fmt.Errorf("failed to load storage items for account %d: %w", accountID, qerr)
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
			continue
		}
		if creator.Valid {
			it.creatorName = creator.String
		}

		if it.slotID <= 0 || int(it.slotID) > len(s.items) {
			continue
		}
		idx := int(it.slotID - 1)
		s.items[idx] = it
		if it.dbID != 0 && it.ID != 0 {
			s.totalSlotsUsed++
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error while reading storage items for account %d: %w", accountID, err)
	}

	return nil
}

func (s *storage) save(accountID int32) (err error) {
	tx, terr := common.DB.Begin()
	if terr != nil {
		return fmt.Errorf("couldn't open transaction to save storage (acct %d): %w", accountID, terr)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, uerr := tx.Exec(
		"UPDATE account_storage SET slots=?, mesos=? WHERE accountID=?",
		s.maxSlots, s.mesos, accountID,
	); uerr != nil {
		err = fmt.Errorf("failed to update storage header (acct %d): %w", accountID, uerr)
		return
	}

	if _, derr := tx.Exec(
		"DELETE FROM account_storage_items WHERE accountID=?",
		accountID,
	); derr != nil {
		err = fmt.Errorf("failed to clear storage items (acct %d): %w", accountID, derr)
		return
	}

	const ins = `
		INSERT INTO account_storage_items(
			accountID, itemID, inventoryID, slotNumber, amount, flag, upgradeSlots, level,
			str, dex, intt, luk, hp, mp, watk, matk, wdef, mdef, accuracy, avoid, hands,
			speed, jump, expireTime, creatorName
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`
	stmt, perr := tx.Prepare(ins)
	if perr != nil {
		err = fmt.Errorf("failed to prepare item insert (acct %d): %w", accountID, perr)
		return
	}
	defer stmt.Close()

	type rowKey struct{ slot int }
	keys := make([]rowKey, 0, len(s.items))
	for i := range s.items {
		keys = append(keys, rowKey{slot: i})
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].slot < keys[j].slot })

	written := 0
	for _, k := range keys {
		it := s.items[k.slot]
		if it.ID == 0 || it.amount == 0 {
			continue
		}

		slotNumber := int16(k.slot + 1)
		if _, ierr := stmt.Exec(
			accountID, it.ID, it.invID, slotNumber, it.amount, it.flag, it.upgradeSlots, it.scrollLevel,
			it.str, it.dex, it.intt, it.luk, it.hp, it.mp, it.watk, it.matk, it.wdef, it.mdef, it.accuracy, it.avoid, it.hands,
			it.speed, it.jump, it.expireTime, it.creatorName,
		); ierr != nil {
			err = fmt.Errorf("failed inserting item %d (acct %d, slot %d): %w", it.ID, accountID, slotNumber, ierr)
			return
		}
		written++
	}

	if cerr := tx.Commit(); cerr != nil {
		err = fmt.Errorf("failed to commit storage save (acct %d): %w", accountID, cerr)
		return
	}

	return nil
}

func (s *storage) addItem(it Item) bool {
	for i := 0; i < int(s.maxSlots); i++ {
		if s.items[i].ID != 0 {
			continue
		}
		s.totalSlotsUsed++
		it.slotID = int16(i + 1)
		s.items[i] = it
		return true
	}
	return false
}

func (s *storage) getAllItems() []Item {
	out := make([]Item, 0, s.totalSlotsUsed)
	for i := range s.items {
		if s.items[i].ID != 0 {
			out = append(out, s.items[i])
		}
	}
	return out
}

func (s *storage) removeAt(idx byte) {
	if idx >= s.maxSlots {
		return
	}
	if s.items[idx].ID == 0 {
		return
	}
	newArr := make([]Item, s.maxSlots)
	dst := 0
	for i := 0; i < int(s.maxSlots); i++ {
		if i == int(idx) {
			continue
		}
		if s.items[i].ID != 0 {
			item := s.items[i]
			dst++
			item.slotID = int16(dst)
			newArr[dst-1] = item
		}
	}
	s.items = newArr
	if s.totalSlotsUsed > 0 {
		s.totalSlotsUsed--
	}
}

func (s *storage) slotsAvailable() bool {
	return s.totalSlotsUsed < s.maxSlots
}

func (s *storage) changeMesos(delta int32) error {
	if delta < 0 && s.mesos < -delta {
		return fmt.Errorf("insufficient storage mesos: have %d, need %d", s.mesos, -delta)
	}
	s.mesos += delta
	return nil
}

func (s *storage) getItemsInSection(inv byte) []Item {
	out := make([]Item, 0, s.totalSlotsUsed)
	for i := range s.items {
		it := s.items[i]
		if it.ID != 0 && it.invID == inv {
			out = append(out, it)
		}
	}
	return out
}

func (s *storage) getBySectionSlot(inv, slot byte) (int, *Item) {
	if slot >= s.maxSlots {
		return -1, nil
	}
	idxInSection := 0
	for i := range s.items {
		if s.items[i].ID == 0 || s.items[i].invID != inv {
			continue
		}
		if idxInSection == int(slot) {
			return i, &s.items[i]
		}
		idxInSection++
	}
	return -1, nil
}
