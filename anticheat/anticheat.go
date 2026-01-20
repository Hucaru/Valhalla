package anticheat

import (
	"database/sql"
	"fmt"
	"time"
)

type AntiCheat struct {
	violations map[string][]time.Time
	failedAuth map[string][]time.Time
	db         *sql.DB
	dispatch   chan func()
	onBan      func(accountID int32, reason string)
}

func New(db *sql.DB, dispatch chan func()) *AntiCheat {
	return &AntiCheat{
		violations: make(map[string][]time.Time),
		failedAuth: make(map[string][]time.Time),
		db:         db,
		dispatch:   dispatch,
	}
}

func (ac *AntiCheat) SetOnBan(fn func(accountID int32, reason string)) {
	ac.onBan = fn
}

// post dispatches function to server's main loop
func (ac *AntiCheat) post(fn func()) {
	if ac.dispatch != nil {
		select {
		case ac.dispatch <- fn:
			return
		default:
			fn()
			return
		}
	}
	fn()
}

// StartCleanup starts periodic cleanup of old violations/auth entries
func (ac *AntiCheat) StartCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			ac.post(func() {
				cutoff := time.Now().Add(-1 * time.Hour)

				for k, timestamps := range ac.violations {
					var keep []time.Time
					for _, t := range timestamps {
						if t.After(cutoff) {
							keep = append(keep, t)
						}
					}
					if len(keep) > 0 {
						ac.violations[k] = keep
					} else {
						delete(ac.violations, k)
					}
				}

				for k, timestamps := range ac.failedAuth {
					var keep []time.Time
					for _, t := range timestamps {
						if t.After(cutoff) {
							keep = append(keep, t)
						}
					}
					if len(keep) > 0 {
						ac.failedAuth[k] = keep
					} else {
						delete(ac.failedAuth, k)
					}
				}
			})
		}
	}()
}

// Track a violation - returns true if threshold exceeded and player should be banned
func (ac *AntiCheat) Track(accountID int32, violationType string, threshold int, window time.Duration) bool {
	key := fmt.Sprintf("%d:%s", accountID, violationType)
	exceeded := false

	ac.post(func() {
		now := time.Now()
		cutoff := now.Add(-window)

		timestamps := []time.Time{now}
		for _, t := range ac.violations[key] {
			if t.After(cutoff) {
				timestamps = append(timestamps, t)
			}
		}
		ac.violations[key] = timestamps
		exceeded = len(timestamps) >= threshold
	})

	return exceeded
}

// Track failed auth attempt - returns true if should ban
func (ac *AntiCheat) TrackFailedAuth(identifier string) bool {
	shouldBan := false

	ac.post(func() {
		now := time.Now()
		cutoff := now.Add(-30 * time.Minute)

		timestamps := []time.Time{now}
		for _, t := range ac.failedAuth[identifier] {
			if t.After(cutoff) {
				timestamps = append(timestamps, t)
			}
		}
		ac.failedAuth[identifier] = timestamps
		shouldBan = len(timestamps) >= 10
	})

	return shouldBan
}

// Clear auth attempts on successful login
func (ac *AntiCheat) ClearAuth(identifiers ...string) {
	ac.post(func() {
		for _, id := range identifiers {
			delete(ac.failedAuth, id)
		}
	})
}

// IssueBan creates a temporary ban (hours=0 means permanent)
func (ac *AntiCheat) IssueBan(accountID int32, hours int, reason string, ip, hwid string) error {
	var banEnd interface{}
	if hours > 0 {
		banEnd = time.Now().Add(time.Duration(hours) * time.Hour)
	}

	_, err := ac.db.Exec(`INSERT INTO bans (accountID, reason, banEnd, ip, hwid) VALUES (?, ?, ?, ?, ?)`,
		accountID, reason, banEnd, ip, hwid)
	if err != nil {
		return err
	}

	_, err = ac.db.Exec(`UPDATE accounts SET isBanned = 1 WHERE accountID = ?`, accountID)

	if ac.onBan != nil {
		ac.post(func() {
			if ac.onBan != nil {
				ac.onBan(accountID, reason)
			}
		})
	}

	if hours > 0 {
		count, _ := ac.incrementTempBans(accountID)
		if count >= 3 {
			err := ac.IssueBan(accountID, 0, "Escalated: 3+ temporary bans", ip, hwid)
			if err != nil {
				return err
			}
		}
	}

	return err
}

// IssueIPBan creates a temporary ban (hours=0 means permanent) for an IP.
func (ac *AntiCheat) IssueIPBan(ip string, hours int, reason string) error {
	var banEnd interface{}
	if hours > 0 {
		banEnd = time.Now().Add(time.Duration(hours) * time.Hour)
	}

	// accountID is intentionally NULL for IP-only bans
	_, err := ac.db.Exec(`INSERT INTO bans (accountID, reason, banEnd, ip, hwid) VALUES (?, ?, ?, ?, ?)`,
		nil, reason, banEnd, ip, "")
	return err
}

// IssueHWIDBan creates a temporary ban (hours=0 means permanent) for an HWID.
func (ac *AntiCheat) IssueHWIDBan(hwid string, hours int, reason string) error {
	var banEnd interface{}
	if hours > 0 {
		banEnd = time.Now().Add(time.Duration(hours) * time.Hour)
	}

	// accountID is intentionally NULL for HWID-only bans
	_, err := ac.db.Exec(`INSERT INTO bans (accountID, reason, banEnd, ip, hwid) VALUES (?, ?, ?, ?, ?)`,
		nil, reason, banEnd, "", hwid)
	return err
}

func (ac *AntiCheat) IsBanned(accountID int32, ip, hwid string) (bool, string, error) {
	var reason string
	err := ac.db.QueryRow(`
SELECT reason FROM bans 
WHERE (accountID = ? OR ip = ? OR (hwid = ? AND hwid != '')) 
AND (banEnd IS NULL OR banEnd > NOW())
LIMIT 1`, accountID, ip, hwid).Scan(&reason)

	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, reason, nil
}

// Unban removes all bans for an account
func (ac *AntiCheat) Unban(accountID int32) error {
	_, err := ac.db.Exec(`DELETE FROM bans WHERE accountID = ?`, accountID)
	if err != nil {
		return err
	}

	_, err = ac.db.Exec(`UPDATE accounts SET isBanned = 0 WHERE accountID = ?`, accountID)
	return err
}

// GetBanHistory returns recent ban records
func (ac *AntiCheat) GetBanHistory(accountID int32, limit int) ([]string, error) {
	rows, err := ac.db.Query(`
SELECT reason, banEnd, createdAt FROM bans 
WHERE accountID = ? 
ORDER BY createdAt DESC LIMIT ?`, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []string
	for rows.Next() {
		var reason string
		var banEnd sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(&reason, &banEnd, &createdAt); err != nil {
			continue
		}

		durStr := "permanent"
		if banEnd.Valid {
			durStr = banEnd.Time.Format("2006-01-02 15:04")
		}
		history = append(history, fmt.Sprintf("%s: %s (until %s)",
			createdAt.Format("2006-01-02 15:04"), reason, durStr))
	}
	return history, nil
}

// Internal: track temp ban count for escalation
func (ac *AntiCheat) incrementTempBans(accountID int32) (int, error) {
	_, err := ac.db.Exec(`
		INSERT INTO ban_escalation (accountID, count) VALUES (?, 1)
		ON DUPLICATE KEY UPDATE count = count + 1`, accountID)
	if err != nil {
		return 0, err
	}

	var count int
	err = ac.db.QueryRow(`SELECT count FROM ban_escalation WHERE accountID = ?`, accountID).Scan(&count)
	return count, err
}

// Detection helpers - track violations and auto-ban on threshold
func (ac *AntiCheat) CheckDamage(accountID int32, damage, maxDamage int32) {
	if damage > maxDamage*2 {
		if ac.Track(accountID, "damage", 5, 5*time.Minute) {
			ac.IssueBan(accountID, 168, fmt.Sprintf("Excessive damage: %d > %d", damage, maxDamage), "", "")
		}
	}
}

func (ac *AntiCheat) CheckAttackSpeed(accountID int32) bool {
	return ac.Track(accountID, "attack_speed", 120, 1*time.Minute)
}

func (ac *AntiCheat) CheckMovement(accountID int32, distance int16) {
	if distance > 1000 {
		if ac.Track(accountID, "teleport", 3, 5*time.Minute) {
			ac.IssueBan(accountID, 168, fmt.Sprintf("Teleport hack: %d pixels", distance), "", "")
		}
	}
}

func (ac *AntiCheat) CheckInvalidItem(accountID int32) {
	if ac.Track(accountID, "invalid_item", 5, 5*time.Minute) {
		ac.IssueBan(accountID, 168, "Using items not in inventory", "", "")
	}
}

func (ac *AntiCheat) CheckInvalidTrade(accountID int32, reason string) {
	if ac.Track(accountID, "invalid_trade", 5, 5*time.Minute) {
		ac.IssueBan(accountID, 168, "Invalid trade: "+reason, "", "")
	}
}

func (ac *AntiCheat) CheckSkillAbuse(accountID int32, skillID int32) {
	if ac.Track(accountID, "skill_abuse", 5, 5*time.Minute) {
		ac.IssueBan(accountID, 168, fmt.Sprintf("Skill abuse: ID %d", skillID), "", "")
	}
}
