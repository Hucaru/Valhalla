package channel

import (
	"database/sql"
	"log"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

const (
	fameErrIncorrectUser byte = 0x01
	fameErrUnderLevel    byte = 0x02
	fameErrThisDay       byte = 0x03
	fameErrThisMonth     byte = 0x04
)

func fameHasRecentActivity(fromID int32, window time.Duration) bool {
	var ts time.Time
	row := common.DB.QueryRow(
		"SELECT `time` FROM `fame_log` WHERE `from`=? AND `time` > (NOW() - INTERVAL ? SECOND) ORDER BY `time` DESC LIMIT 1",
		fromID, int64(window.Seconds()),
	)
	switch err := row.Scan(&ts); err {
	case nil:
		return true
	case sql.ErrNoRows:
		return false
	default:
		log.Println("fameHasRecentActivity:", err)
		return true
	}
}

func fameInsertLog(fromID, toID int32) error {
	_, err := common.DB.Exec(
		"INSERT INTO fame_log (`from`, `to`, `time`) VALUES (?,?, NOW())",
		fromID, toID,
	)
	return err
}

func packetFameError(code byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelFameOperation)
	p.WriteByte(code)
	return p
}

func packetFameNotifyVictim(fromName string, up bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelFameOperation)
	p.WriteByte(0x05)
	p.WriteString(fromName)
	p.WriteBool(up)
	return p
}

func packetFameNotifySource(victimName string, up bool, newFame int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelFameOperation)
	p.WriteByte(0x00)
	p.WriteString(victimName)
	p.WriteBool(up)
	p.WriteInt32(int32(newFame))
	return p
}
