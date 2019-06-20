package game

import (
	"fmt"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
)

type players []*player

func (p *players) getFromConn(conn mnet.Client) (*player, error) {
	for _, v := range *p {
		if v.conn == conn {
			return v, nil
		}
	}

	return nil, fmt.Errorf("Could not retrieve player")
}

func (p *players) removeFromConn(conn mnet.Client) error {
	i := -1

	for j, v := range *p {
		if v.conn == conn {
			i = j
			break
		}
	}

	if i == -1 {
		return fmt.Errorf("Could not find player")
	}

	(*p)[i] = (*p)[len(*p)-1]
	(*p)[len(*p)-1] = nil
	(*p) = (*p)[:len(*p)-1]

	return nil
}

type player struct {
	conn       mnet.Client
	char       character
	instanceID int
}

func newPlayer(conn mnet.Client, char character) *player {
	return &player{conn: conn, char: char, instanceID: 0}
}

func (p *player) setJob(id int16) {
	p.char.job = id
	p.conn.Send(packetPlayerStatChange(true, constant.JobID, int32(id)))
}

func (p *player) setLevel(amount byte) {

}

func (p *player) giveLevel(amount byte) {

}

func (p *player) setStr(amount int16) {

}

func (p *player) setDex(amount int16) {

}

func (p *player) setInt(amount int16) {

}

func (p *player) setLuk(amount int16) {

}

func (p *player) setHP(amount int16) {

}

func (p *player) setMaxHP(amount int16) {

}

func (p *player) setMP(amount int16) {

}

func (p *player) setMaxMP(amount int16) {

}

func (p *player) setAP(amount int16) {

}

func (p *player) setSp(amount int16) {

}

func (p *player) setEXP(amount int32) {

}

func (p *player) giveEXP(amount int32) {

}

func (p *player) setFame(amount int16) {

}

func (p *player) setGuild(name string) {

}

func (p *player) setEquipSlotSize(size byte) {

}

func (p *player) setUseSlotSize(size byte) {

}

func (p *player) setEtcSlotSize(size byte) {

}

func (p *player) setCashSlotSize(size byte) {

}

func (p *player) setMesos(amount int32) {

}

func (p *player) giveMesos(amount int32) {

}

func (p *player) setMinigameWins(amount int32) {

}

func (p *player) setMinigameDraws(amount int32) {

}

func (p *player) setMinigameLoss(amount int32) {

}

func (p *player) updateMovement(frag movementFrag) {
	p.char.pos.x = frag.x
	p.char.pos.y = frag.y
	p.char.foothold = frag.foothold
	p.char.stance = frag.stance
}
