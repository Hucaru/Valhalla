package channel

import (
	"database/sql"
	"fmt"

	"github.com/Hucaru/Valhalla/common"
)

type shopEscrowRow struct {
	ID           int64
	OwnerID      int32
	ShopSlot     byte
	Price        int32
	Bundles      int16
	BundleAmount int16
	Item         Item
}

func shopEscrowInsert(ownerID int32, shopSlot byte, price int32, bundles, bundleAmount int16, item Item) (int64, error) {
	res, err := common.DB.Exec(`
		INSERT INTO player_shop_escrow_items(
			ownerCharacterID, shopSlot, price, bundles, bundleAmount,
			itemID, inventoryID, amount, flag, upgradeSlots, level,
			str, dex, intt, luk, hp, mp, watk, matk, wdef, mdef, accuracy, avoid, hands, speed, jump,
			expireTime, creatorName, cashID, cashSN
		) VALUES (?,?,?,?,?,
			?,?,?,?,?,?,
			?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,
			?,?,?,?
		)`,
		ownerID, shopSlot, price, bundles, bundleAmount,
		item.ID, item.invID, item.amount, item.flag, item.upgradeSlots, item.scrollLevel,
		item.str, item.dex, item.intt, item.luk, item.hp, item.mp, item.watk, item.matk, item.wdef, item.mdef, item.accuracy, item.avoid, item.hands, item.speed, item.jump,
		item.expireTime, item.creatorName, sql.NullInt64{Int64: item.cashID, Valid: item.cashID != 0}, sql.NullInt32{Int32: int32(item.cashSN), Valid: item.cashSN != 0},
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func shopEscrowUpdateBundles(escrowID int64, bundles int16, newAmount int16) error {
	_, err := common.DB.Exec(`
		UPDATE player_shop_escrow_items
		SET bundles=?, amount=?
		WHERE id=?`,
		bundles, newAmount, escrowID,
	)
	return err
}

func shopEscrowDelete(escrowID int64) error {
	_, err := common.DB.Exec(`DELETE FROM player_shop_escrow_items WHERE id=?`, escrowID)
	return err
}

func shopEscrowLoadByOwner(ownerID int32) ([]shopEscrowRow, error) {
	rows, err := common.DB.Query(`
		SELECT
			id, ownerCharacterID, shopSlot, price, bundles, bundleAmount,
			itemID, inventoryID, amount, flag, upgradeSlots, level,
			str, dex, intt, luk, hp, mp, watk, matk, wdef, mdef, accuracy, avoid, hands, speed, jump,
			expireTime, creatorName, cashID, cashSN
		FROM player_shop_escrow_items
		WHERE ownerCharacterID=?
		ORDER BY id ASC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]shopEscrowRow, 0, 8)

	for rows.Next() {
		var r shopEscrowRow
		var cashID sql.NullInt64
		var cashSN sql.NullInt32

		if err := rows.Scan(
			&r.ID, &r.OwnerID, &r.ShopSlot, &r.Price, &r.Bundles, &r.BundleAmount,
			&r.Item.ID, &r.Item.invID, &r.Item.amount, &r.Item.flag, &r.Item.upgradeSlots, &r.Item.scrollLevel,
			&r.Item.str, &r.Item.dex, &r.Item.intt, &r.Item.luk, &r.Item.hp, &r.Item.mp, &r.Item.watk, &r.Item.matk,
			&r.Item.wdef, &r.Item.mdef, &r.Item.accuracy, &r.Item.avoid, &r.Item.hands, &r.Item.speed, &r.Item.jump,
			&r.Item.expireTime, &r.Item.creatorName, &cashID, &cashSN,
		); err != nil {
			return nil, err
		}

		if cashID.Valid {
			r.Item.cashID = cashID.Int64
		}
		if cashSN.Valid {
			r.Item.cashSN = int32(cashSN.Int32)
		}

		r.Item.dbID = 0
		r.Item.slotID = 0

		out = append(out, r)
	}

	return out, rows.Err()
}

func shopEscrowMesosAdd(ownerID int32, delta int32) error {
	if delta <= 0 {
		return nil
	}
	_, err := common.DB.Exec(`
		INSERT INTO player_shop_escrow_mesos(ownerCharacterID, mesos)
		VALUES(?, ?)
		ON DUPLICATE KEY UPDATE mesos = mesos + VALUES(mesos)`,
		ownerID, delta,
	)
	return err
}

func shopEscrowMesosClaim(ownerID int32) (int32, error) {
	var mesos int32
	err := common.DB.QueryRow(`SELECT mesos FROM player_shop_escrow_mesos WHERE ownerCharacterID=?`, ownerID).Scan(&mesos)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	_, _ = common.DB.Exec(`DELETE FROM player_shop_escrow_mesos WHERE ownerCharacterID=?`, ownerID)
	return mesos, nil
}

// restoreShopEscrowOnLogin returns any escrowed shop items + banked mesos to the player
func (d *Player) restoreShopEscrowOnLogin() {
	if d == nil {
		return
	}

	items, err := shopEscrowLoadByOwner(d.ID)
	if err == nil && len(items) > 0 {
		for _, row := range items {
			ret := row.Item
			if row.Bundles > 0 && row.BundleAmount > 0 {
				ret.amount = row.Bundles * row.BundleAmount
			}
			ret.dbID = 0
			ret.slotID = 0

			_, _ = d.GiveItem(ret)

			_ = shopEscrowDelete(row.ID)
		}
	}

	mesos, err := shopEscrowMesosClaim(d.ID)
	if err == nil && mesos > 0 {
		d.mesos += mesos
		_, _ = common.DB.Exec(`UPDATE characters SET mesos=? WHERE ID=?`, d.mesos, d.ID)
	}
}

func shopEscrowAmountSanity(bundles, bundleAmount int16) error {
	if bundles <= 0 || bundleAmount <= 0 {
		return fmt.Errorf("invalid bundles/bundleAmount")
	}
	return nil
}
