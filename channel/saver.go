package channel

import (
	"log"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/common"
)

// DirtyBits mark which character columns need persisting.
type DirtyBits uint64

const (
	DirtyAP DirtyBits = 1 << iota
	DirtySP
	DirtyMesos
	DirtyHP
	DirtyMP
	DirtyMaxHP
	DirtyMaxMP
	DirtyEXP
	DirtyMap
	DirtyPrevMap
	DirtyJob
	DirtyLevel
	DirtyStr
	DirtyDex
	DirtyInt
	DirtyLuk
	DirtyFame
	DirtyInvSlotSizes
	DirtyMiniGame
	DirtyBuddySize
	DirtySkills
	DirtyNX
	DirtyMaplePoints
)

// snapshot contains only columns we may persist.
type snapshot struct {
	ID        int32
	AccountID int32

	AP, SP                 int16
	Mesos, NX, MaplePoints int32
	HP, MaxHP              int16
	MP, MaxMP              int16
	EXP                    int32
	MapID                  int32
	PrevMapID              int32
	MapPos                 byte
	Job                    int16
	Level                  byte
	Str, Dex, Intt         int16
	Luk, Fame              int16

	EquipSlotSize byte
	UseSlotSize   byte
	SetupSlotSize byte
	EtcSlotSize   byte
	CashSlotSize  byte

	MiniGameWins   int32
	MiniGameDraw   int32
	MiniGameLoss   int32
	MiniGamePoints int32

	BuddyListSize byte

	Skills map[int32]playerSkill
}

type pendingSave struct {
	due  time.Time
	bits DirtyBits
	snap snapshot
}

type saver struct {
	schedCh chan scheduleReq
	flushCh chan flushReq
	stopCh  chan struct{}

	pending map[int32]*pendingSave
	ticker  *time.Ticker
}

var saverInst *saver

// Requests
type scheduleReq struct {
	id    int32
	bits  DirtyBits
	snap  snapshot
	delay time.Duration
}

type flushReq struct {
	id int32
	// If set, use this bits/snap instead of pending/live reconcile.
	overrideBits DirtyBits
	overrideSnap *snapshot
	done         chan struct{} // closed when flush completes
}

// Helpers

func copySkills(src map[int32]playerSkill) map[int32]playerSkill {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[int32]playerSkill, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// Build a snapshot from a Player (caller is on game thread).
func snapshotFromPlayer(p *Player) snapshot {
	s := snapshot{
		ID:             p.ID,
		AccountID:      p.accountID,
		AP:             p.ap,
		SP:             p.sp,
		Mesos:          p.mesos,
		NX:             p.nx,
		MaplePoints:    p.maplepoints,
		HP:             p.hp,
		MaxHP:          p.maxHP,
		MP:             p.mp,
		MaxMP:          p.maxMP,
		EXP:            p.exp,
		MapID:          p.mapID,
		PrevMapID:      p.previousMap,
		MapPos:         p.mapPos,
		Job:            p.job,
		Level:          p.level,
		Str:            p.str,
		Dex:            p.dex,
		Intt:           p.intt,
		Luk:            p.luk,
		Fame:           p.fame,
		EquipSlotSize:  p.equipSlotSize,
		UseSlotSize:    p.useSlotSize,
		SetupSlotSize:  p.setupSlotSize,
		EtcSlotSize:    p.etcSlotSize,
		CashSlotSize:   p.cashSlotSize,
		MiniGameWins:   p.miniGameWins,
		MiniGameDraw:   p.miniGameDraw,
		MiniGameLoss:   p.miniGameLoss,
		MiniGamePoints: p.miniGamePoints,
		BuddyListSize:  p.buddyListSize,
	}

	if p.dirty&DirtySkills != 0 {
		s.Skills = copySkills(p.skills)
	}
	return s
}

func init() {
	if saverInst == nil {
		saverInst = &saver{
			schedCh: make(chan scheduleReq, 1024),
			flushCh: make(chan flushReq, 256),
			stopCh:  make(chan struct{}),
			pending: make(map[int32]*pendingSave),
			ticker:  time.NewTicker(50 * time.Millisecond),
		}
		go saverInst.loop()
	}
}

func StopSaver() {
	if saverInst == nil {
		return
	}
	close(saverInst.stopCh)
	saverInst.ticker.Stop()
	saverInst = nil
}

func scheduleSave(p *Player, delay time.Duration) {
	if saverInst == nil || p == nil || p.ID == 0 {
		return
	}
	req := scheduleReq{
		id:    p.ID,
		bits:  p.dirty,
		snap:  snapshotFromPlayer(p),
		delay: delay,
	}
	select {
	case saverInst.schedCh <- req:
	default:
		// drop under extreme pressure
	}
}

func FlushNow(p *Player) {
	if saverInst == nil || p == nil || p.ID == 0 {
		return
	}
	done := make(chan struct{})
	var snap *snapshot
	var bits DirtyBits
	if p.dirty != 0 {
		sn := snapshotFromPlayer(p)
		snap = &sn
		bits = p.dirty
	}
	req := flushReq{
		id:           p.ID,
		overrideBits: bits,
		overrideSnap: snap,
		done:         done,
	}
	select {
	case saverInst.flushCh <- req:
		<-done
	default:
		// channel saturated: fallback to direct sync persist
		job := pendingSave{bits: bits, snap: snapshotFromPlayer(p)}
		saverInst.persist(job)
	}
}

func (s *saver) loop() {
	for {
		select {
		case <-s.stopCh:
			return

		case req := <-s.schedCh:
			ps := s.pending[req.id]
			now := time.Now()
			due := now.Add(req.delay)
			if ps == nil {
				ps = &pendingSave{due: due, bits: req.bits, snap: req.snap}
				s.pending[req.id] = ps
			} else {
				ps.bits |= req.bits
				mergeSnapshot(&ps.snap, req.snap)
				if due.Before(ps.due) {
					ps.due = due
				}
			}

		case req := <-s.flushCh:
			ps, ok := s.pending[req.id]
			var job pendingSave
			if ok {
				job = *ps
				delete(s.pending, req.id)
			} else {
				job = pendingSave{bits: 0, snap: snapshot{ID: req.id}}
			}
			if req.overrideSnap != nil {
				job.snap = *req.overrideSnap
				job.bits |= req.overrideBits
			}
			s.persist(job)
			if req.done != nil {
				close(req.done)
			}

		case <-s.ticker.C:
			now := time.Now()
			var dueList []pendingSave
			for id, ps := range s.pending {
				if now.After(ps.due) {
					dueList = append(dueList, *ps)
					delete(s.pending, id)
				}
			}
			for _, job := range dueList {
				go s.persist(job)
			}
		}
	}
}

// Merge two snapshots (in-place) using rhs as the latest authoritative values.
func mergeSnapshot(lhs *snapshot, rhs snapshot) {
	lhs.AP, lhs.SP = rhs.AP, rhs.SP
	lhs.Mesos = rhs.Mesos
	lhs.NX = rhs.NX
	lhs.MaplePoints = rhs.MaplePoints
	lhs.HP, lhs.MaxHP = rhs.HP, rhs.MaxHP
	lhs.MP, lhs.MaxMP = rhs.MP, rhs.MaxMP
	lhs.EXP = rhs.EXP
	lhs.MapID, lhs.MapPos = rhs.MapID, rhs.MapPos
	lhs.PrevMapID = rhs.PrevMapID
	lhs.Job, lhs.Level = rhs.Job, rhs.Level
	lhs.Str, lhs.Dex, lhs.Intt, lhs.Luk, lhs.Fame = rhs.Str, rhs.Dex, rhs.Intt, rhs.Luk, rhs.Fame

	lhs.EquipSlotSize = rhs.EquipSlotSize
	lhs.UseSlotSize = rhs.UseSlotSize
	lhs.SetupSlotSize = rhs.SetupSlotSize
	lhs.EtcSlotSize = rhs.EtcSlotSize
	lhs.CashSlotSize = rhs.CashSlotSize

	lhs.MiniGameWins = rhs.MiniGameWins
	lhs.MiniGameDraw = rhs.MiniGameDraw
	lhs.MiniGameLoss = rhs.MiniGameLoss
	lhs.MiniGamePoints = rhs.MiniGamePoints

	lhs.BuddyListSize = rhs.BuddyListSize

	if rhs.Skills != nil {
		lhs.Skills = rhs.Skills
	}
}

func (s *saver) persist(job pendingSave) bool {
	if job.bits == 0 {
		return true
	}

	cols := make([]string, 0, 24)
	args := make([]any, 0, 24)

	if job.bits&DirtyAP != 0 {
		cols = append(cols, "ap=?")
		args = append(args, job.snap.AP)
	}
	if job.bits&DirtySP != 0 {
		cols = append(cols, "sp=?")
		args = append(args, job.snap.SP)
	}
	if job.bits&DirtyMesos != 0 {
		cols = append(cols, "mesos=?")
		args = append(args, job.snap.Mesos)
	}
	if job.bits&DirtyHP != 0 {
		cols = append(cols, "hp=?")
		args = append(args, job.snap.HP)
	}
	if job.bits&DirtyMaxHP != 0 {
		cols = append(cols, "maxHP=?")
		args = append(args, job.snap.MaxHP)
	}
	if job.bits&DirtyMP != 0 {
		cols = append(cols, "mp=?")
		args = append(args, job.snap.MP)
	}
	if job.bits&DirtyMaxMP != 0 {
		cols = append(cols, "maxMP=?")
		args = append(args, job.snap.MaxMP)
	}
	if job.bits&DirtyEXP != 0 {
		cols = append(cols, "exp=?")
		args = append(args, job.snap.EXP)
	}
	if job.bits&DirtyMap != 0 {
		cols = append(cols, "mapID=?, mapPos=?")
		args = append(args, job.snap.MapID, job.snap.MapPos)
	}
	if job.bits&DirtyPrevMap != 0 {
		cols = append(cols, "previousMapID=?")
		args = append(args, job.snap.PrevMapID)
	}
	if job.bits&DirtyJob != 0 {
		cols = append(cols, "job=?")
		args = append(args, job.snap.Job)
	}
	if job.bits&DirtyLevel != 0 {
		cols = append(cols, "level=?")
		args = append(args, job.snap.Level)
	}
	if job.bits&DirtyStr != 0 {
		cols = append(cols, "str=?")
		args = append(args, job.snap.Str)
	}
	if job.bits&DirtyDex != 0 {
		cols = append(cols, "dex=?")
		args = append(args, job.snap.Dex)
	}
	if job.bits&DirtyInt != 0 {
		cols = append(cols, "intt=?")
		args = append(args, job.snap.Intt)
	}
	if job.bits&DirtyLuk != 0 {
		cols = append(cols, "luk=?")
		args = append(args, job.snap.Luk)
	}
	if job.bits&DirtyFame != 0 {
		cols = append(cols, "fame=?")
		args = append(args, job.snap.Fame)
	}
	if job.bits&DirtyInvSlotSizes != 0 {
		cols = append(cols, "equipSlotSize=?,useSlotSize=?,setupSlotSize=?,etcSlotSize=?,cashSlotSize=?")
		args = append(args, job.snap.EquipSlotSize, job.snap.UseSlotSize, job.snap.SetupSlotSize, job.snap.EtcSlotSize, job.snap.CashSlotSize)
	}
	if job.bits&DirtyMiniGame != 0 {
		cols = append(cols, "miniGameWins=?,miniGameDraw=?,miniGameLoss=?,miniGamePoints=?")
		args = append(args, job.snap.MiniGameWins, job.snap.MiniGameDraw, job.snap.MiniGameLoss, job.snap.MiniGamePoints)
	}
	if job.bits&DirtyBuddySize != 0 {
		cols = append(cols, "buddyListSize=?")
		args = append(args, job.snap.BuddyListSize)
	}

	if len(cols) > 0 {
		query := "UPDATE characters SET " + strings.Join(cols, ",") + " WHERE ID=?"
		args = append(args, job.snap.ID)
		if _, err := common.DB.Exec(query, args...); err != nil {
			log.Printf("saver.persist: UPDATE characters (ID=%d) failed: %v", job.snap.ID, err)
		}
	}

	if job.bits&DirtySkills != 0 && len(job.snap.Skills) > 0 {
		upsert := `INSERT INTO skills(characterID,skillID,level,cooldown)
		           VALUES(?,?,?,?)
		           ON DUPLICATE KEY UPDATE level=VALUES(level), cooldown=VALUES(cooldown)`
		for sid, srec := range job.snap.Skills {
			if _, err := common.DB.Exec(upsert, job.snap.ID, sid, srec.Level, srec.Cooldown); err != nil {
				log.Printf("saver.persist: upsert skill %d for char %d failed: %v", sid, job.snap.ID, err)
			}
		}
	}

	if job.bits&DirtyNX != 0 || job.bits&DirtyMaplePoints != 0 {
		query := "UPDATE accounts SET nx=?, maplepoints=? WHERE accountID=?"
		if _, err := common.DB.Exec(query, job.snap.NX, job.snap.MaplePoints, job.snap.AccountID); err != nil {
			log.Printf("saver.persist: UPDATE accounts (ID=%d) failed: %v", job.snap.AccountID, err)
		}
	}

	return true
}
