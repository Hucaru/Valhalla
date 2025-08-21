package channel

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/common"
)

// DirtyBits mark which character columns need persisting.
type DirtyBits uint32

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
	DirtyJob
	DirtyLevel
	DirtyStr
	DirtyDex
	DirtyInt
	DirtyLuk
	DirtyFame
)

// snapshot contains only columns we may persist.
type snapshot struct {
	ID int32

	AP, SP         int16
	Mesos          int32
	HP, MaxHP      int16
	MP, MaxMP      int16
	EXP            int32
	MapID          int32
	MapPos         byte
	Job            int16
	Level          byte
	Str, Dex, Intt int16
	Luk, Fame      int16
}

// pendingSave is a coalesced, debounced update for one character.
type pendingSave struct {
	due  time.Time
	bits DirtyBits
	snap snapshot
}

// Saver batches and persists character updates.
type Saver struct {
	mu      sync.Mutex
	pending map[int32]*pendingSave // charID -> pendingSave
	ticker  *time.Ticker
	stopCh  chan struct{}
}

// SaverInstance is the global saver used by the channel server.
var SaverInstance *Saver

// StartSaver boots the saver loop; should be called once at startup.
func StartSaver() {
	if SaverInstance != nil {
		return
	}
	SaverInstance = &Saver{
		pending: make(map[int32]*pendingSave),
		ticker:  time.NewTicker(50 * time.Millisecond),
		stopCh:  make(chan struct{}),
	}
	go SaverInstance.loop()
}

// StopSaver stops the saver (e.g., on shutdown).
func StopSaver() {
	if SaverInstance == nil {
		return
	}
	close(SaverInstance.stopCh)
	SaverInstance.ticker.Stop()
	SaverInstance = nil
}

// SchedulePlayer copies the player’s current values and schedules a debounced write.
// delay is the debounce window to coalesce multiple changes.
func (s *Saver) SchedulePlayer(p *player, delay time.Duration) {
	if p == nil || p.id == 0 {
		return
	}
	now := time.Now()
	s.mu.Lock()
	ps := s.pending[p.id]
	if ps == nil {
		ps = &pendingSave{
			due: now.Add(delay),
			snap: snapshot{
				ID:     p.id,
				AP:     p.ap,
				SP:     p.sp,
				Mesos:  p.mesos,
				HP:     p.hp,
				MaxHP:  p.maxHP,
				MP:     p.mp,
				MaxMP:  p.maxMP,
				EXP:    p.exp,
				MapID:  p.mapID,
				MapPos: p.mapPos,
				Job:    p.job,
				Level:  p.level,
				Str:    p.str,
				Dex:    p.dex,
				Intt:   p.intt,
				Luk:    p.luk,
				Fame:   p.fame,
			},
			bits: p.dirty,
		}
		s.pending[p.id] = ps
	} else {
		// Coalesce: update snapshot to latest and merge dirty bits.
		ps.snap.AP = p.ap
		ps.snap.SP = p.sp
		ps.snap.Mesos = p.mesos
		ps.snap.HP = p.hp
		ps.snap.MaxHP = p.maxHP
		ps.snap.MP = p.mp
		ps.snap.MaxMP = p.maxMP
		ps.snap.EXP = p.exp
		ps.snap.MapID = p.mapID
		ps.snap.MapPos = p.mapPos
		ps.snap.Job = p.job
		ps.snap.Level = p.level
		ps.snap.Str = p.str
		ps.snap.Dex = p.dex
		ps.snap.Intt = p.intt
		ps.snap.Luk = p.luk
		ps.snap.Fame = p.fame

		ps.bits |= p.dirty
		// Move due earlier if needed (sooner flush).
		d := now.Add(delay)
		if d.Before(ps.due) {
			ps.due = d
		}
	}
	// The player's dirty flags remain set until the flush succeeds.
	s.mu.Unlock()
}

// FlushNow persists the pending data for this character synchronously.
func (s *Saver) FlushNow(p *player) {
	if p == nil || p.id == 0 {
		return
	}
	var job *pendingSave
	s.mu.Lock()
	if ps, ok := s.pending[p.id]; ok {
		job = &pendingSave{
			due:  time.Now(),
			bits: ps.bits,
			snap: ps.snap,
		}
		delete(s.pending, p.id)
	} else if p.dirty != 0 {
		// No pending entry, but player has dirty bits; create a job from live values.
		job = &pendingSave{
			due:  time.Now(),
			bits: p.dirty,
			snap: snapshot{
				ID:     p.id,
				AP:     p.ap,
				SP:     p.sp,
				Mesos:  p.mesos,
				HP:     p.hp,
				MaxHP:  p.maxHP,
				MP:     p.mp,
				MaxMP:  p.maxMP,
				EXP:    p.exp,
				MapID:  p.mapID,
				MapPos: p.mapPos,
				Job:    p.job,
				Level:  p.level,
				Str:    p.str,
				Dex:    p.dex,
				Intt:   p.intt,
				Luk:    p.luk,
				Fame:   p.fame,
			},
		}
	}
	s.mu.Unlock()

	if job != nil {
		if s.persist(*job) {
			// Clear player's dirty bits after successful flush.
			p.clearDirty(job.bits)
		}
	}
}

// Internal loop ticking due jobs.
func (s *Saver) loop() {
	for {
		select {
		case <-s.stopCh:
			return
		case <-s.ticker.C:
			s.flushDue()
		}
	}
}

// Move due jobs out and persist them without holding the lock.
func (s *Saver) flushDue() {
	now := time.Now()
	var batch []pendingSave

	s.mu.Lock()
	for id, ps := range s.pending {
		if now.After(ps.due) {
			batch = append(batch, *ps)
			delete(s.pending, id)
		}
	}
	s.mu.Unlock()

	for _, job := range batch {
		if s.persist(job) {
			// Best-effort: clear the player's dirty bits if player is still in memory.
			// We cannot access the player map here without references; callers also clear.
		}
	}
}

// Build a single UPDATE for only the changed columns.
func (s *Saver) persist(job pendingSave) bool {
	if job.bits == 0 {
		return true
	}

	cols := make([]string, 0, 16)
	args := make([]any, 0, 16)

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

	if len(cols) == 0 {
		return true
	}

	query := "UPDATE characters SET " + strings.Join(cols, ",") + " WHERE id=?"
	args = append(args, job.snap.ID)

	if _, err := common.DB.Exec(query, args...); err != nil {
		log.Printf("Saver.persist: UPDATE characters (id=%d) failed: %v", job.snap.ID, err)
		return false
	}
	return true
}
