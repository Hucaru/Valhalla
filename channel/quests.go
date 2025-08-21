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
	inProgress []quest
	completed  []quest
}

type quest struct {
	id          int16
	name        string  // reuse as "record" text for active
	items       []int32 // unused here
	completedAt int64   // Unix ms
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
			q.completed = append(q.completed, quest{id: id, completedAt: completedAt})
		} else {
			q.inProgress = append(q.inProgress, quest{id: id, name: record})
		}
	}
	return q
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

// In-memory helpers
func (q *quests) add(id int16, record string) {
	q.inProgress = append(q.inProgress, quest{id: id, name: record})
}
func (q *quests) remove(id int16) {
	for i, v := range q.inProgress {
		if v.id == id {
			q.inProgress[i] = q.inProgress[len(q.inProgress)-1]
			q.inProgress = q.inProgress[:len(q.inProgress)-1]
			return
		}
	}
}
func (q *quests) complete(id int16, completedAtMs int64) {
	q.completed = append(q.completed, quest{id: id, completedAt: completedAtMs})
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
	p.WriteByte(0x01)
	p.WriteString("")
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
