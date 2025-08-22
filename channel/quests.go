package channel

import (
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

const (
	QUEST_LOST_ITEM = 0x00
	QUEST_STARTED   = 0x01
	QUEST_COMPLETED = 0x02
	QUEST_FORFEIT   = 0x03
)

type quests struct {
	inProgress map[int16]quest
	completed  map[int16]quest

	mobKills map[int16]map[int32]int32
}

type quest struct {
	id          int16
	name        string
	items       []int32
	completedAt int64
}

func newQuests() quests {
	q := quests{}
	q.init()
	return q
}

func (q *quests) init() {
	if q.inProgress == nil {
		q.inProgress = make(map[int16]quest, 16)
	}
	if q.completed == nil {
		q.completed = make(map[int16]quest, 16)
	}
	if q.mobKills == nil {
		q.mobKills = make(map[int16]map[int32]int32, 16)
	}
}

// FILETIME helpers
func toFileTime(t time.Time) int64 {
	const ticksPerSecond = int64(10_000_000)
	const unixToFiletimeSeconds = int64(11644473600)
	sec := t.Unix()
	nsec := t.UnixNano() - sec*1_000_000_000
	return (sec+unixToFiletimeSeconds)*ticksPerSecond + nsec/100
}
func unixMsToFileTime(ms int64) int64 {
	return toFileTime(time.Unix(0, ms*int64(time.Millisecond)))
}

// Single-table DB I/O
func loadQuestsFromDB(charID int32) quests {
	var q quests
	q.init() // ensure maps are ready

	rows, err := common.DB.Query(
		"SELECT questID, record, completed, completedAt FROM character_quests WHERE characterID=?",
		charID,
	)
	if err != nil {
		return q
	}
	defer rows.Close()

	for rows.Next() {
		var id int16
		var record string
		var completed bool
		var completedAt int64
		if err := rows.Scan(&id, &record, &completed, &completedAt); err != nil {
			continue
		}
		if completed {
			q.completed[id] = quest{id: id, completedAt: completedAt}
		} else {
			q.inProgress[id] = quest{id: id, name: record}
		}
	}
	return q
}

func loadQuestMobKillsFromDB(charID int32) map[int16]map[int32]int32 {
	out := make(map[int16]map[int32]int32, 16)

	rows, err := common.DB.Query(
		"SELECT questID, mobID, kills FROM character_quest_kills WHERE characterID=?",
		charID,
	)
	if err != nil {
		return out
	}
	defer rows.Close()

	for rows.Next() {
		var qid int16
		var mobID int32
		var kills int32
		if err := rows.Scan(&qid, &mobID, &kills); err != nil {
			continue
		}
		if _, ok := out[qid]; !ok {
			out[qid] = make(map[int32]int32, 4)
		}
		out[qid][mobID] = kills
	}
	return out
}

func upsertQuestRecord(charID int32, questID int16, record string) {
	_, _ = common.DB.Exec(
		"INSERT INTO character_quests(characterID, questID, record, completed, completedAt) "+
			"VALUES(?,?,?,?,?) ON DUPLICATE KEY UPDATE record=VALUES(record), completed=0, completedAt=0",
		charID, questID, record, 0, 0,
	)
}

func setQuestCompleted(charID int32, questID int16, completedAtMs int64) {
	_, _ = common.DB.Exec(
		"INSERT INTO character_quests(characterID, questID, record, completed, completedAt) "+
			"VALUES(?,?,?,?,?) ON DUPLICATE KEY UPDATE completed=1, completedAt=VALUES(completedAt)",
		charID, questID, "", 1, completedAtMs,
	)
}

func deleteQuest(charID int32, questID int16) {
	_, _ = common.DB.Exec("DELETE FROM character_quests WHERE characterID=? AND questID=?", charID, questID)
}

func upsertQuestMobKill(charID int32, questID int16, mobID int32, delta int32) {
	_, _ = common.DB.Exec(
		"INSERT INTO character_quest_kills(characterID, questID, mobID, kills) VALUES(?,?,?,?) "+
			"ON DUPLICATE KEY UPDATE kills = kills + VALUES(kills)",
		charID, questID, mobID, delta,
	)
}

func clearQuestMobKills(charID int32, questID int16) {
	_, _ = common.DB.Exec(
		"DELETE FROM character_quest_kills WHERE characterID=? AND questID=?",
		charID, questID,
	)
}

// In-memory helpers
func (q *quests) add(id int16, name string) {
	q.inProgress[id] = quest{id: id, name: name}
	delete(q.completed, id)

}

func (q *quests) remove(id int16) {
	delete(q.inProgress, id)
}
func (q *quests) complete(id int16, completedAt int64) {
	delete(q.inProgress, id)
	q.completed[id] = quest{id: id, completedAt: completedAt}
	delete(q.mobKills, id)
}

func (q *quests) hasInProgress(id int16) bool {
	_, ok := q.inProgress[id]
	return ok
}

func (q *quests) hasCompleted(id int16) bool {
	_, ok := q.completed[id]
	return ok
}

func (q quests) inProgressList() []quest {
	if len(q.inProgress) == 0 {
		return nil
	}
	out := make([]quest, 0, len(q.inProgress))
	for _, v := range q.inProgress {
		out = append(out, v)
	}
	return out
}

func (q quests) completedList() []quest {
	if len(q.completed) == 0 {
		return nil
	}
	out := make([]quest, 0, len(q.completed))
	for _, v := range q.completed {
		out = append(out, v)
	}
	return out
}

// Runtime update packets
func packetCompleteQuest(questID int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessage)
	p.WriteByte(0x01)
	p.WriteInt16(questID)
	p.WriteByte(0x02)
	p.WriteInt64(toFileTime(time.Now()))
	return p
}
func packetUpdateQuest(questID int16, data string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessage)
	p.WriteByte(0x01)
	p.WriteInt16(questID)
	p.WriteByte(0x01)
	p.WriteString(data)
	return p
}
func packetRemoveQuest(questID int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessage)
	p.WriteByte(0x01)
	p.WriteInt16(questID)
	p.WriteInt16(0)
	p.WriteByte(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	return p
}

func packetUpdateQuestMobKills(questID int16, killStr string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessage)
	p.WriteByte(0x01)
	p.WriteInt16(questID)
	p.WriteByte(0x01)
	p.WriteString(killStr)
	p.WriteInt32(0)
	p.WriteInt32(0)
	return p
}

// Login serialization
func writeActiveQuests(p *mpacket.Packet, qs []quest) {
	p.WriteInt16(int16(len(qs)))
	for _, v := range qs {
		p.WriteInt16(v.id)
		p.WriteString(v.name) // record text
	}
}
func writeCompletedQuests(p *mpacket.Packet, qs []quest) {
	p.WriteInt16(int16(len(qs)))
	for _, v := range qs {
		p.WriteInt16(v.id)
		p.WriteInt64(unixMsToFileTime(v.completedAt))
	}
}
