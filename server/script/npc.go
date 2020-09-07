package script

import "github.com/Hucaru/Valhalla/mnet"

type NpcState struct {
	npcID int32
	state int
	conn  mnet.Client
}

func CreateNewNpcState(npcID int32, conn mnet.Client) *NpcState {
	return &NpcState{npcID: npcID, state: 0, conn: conn}
}

func (n *NpcState) SendBackNext(msg string, back, next bool) {
	n.conn.Send(packetChatBackNext(n.npcID, msg, next, back))
}

func (n *NpcState) SendOK(msg string) {
	n.conn.Send(packetChatOk(n.npcID, msg))
}
