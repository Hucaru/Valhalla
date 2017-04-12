package main

import "github.com/Hucaru/Valhalla/common"

func SendHandshake(client common.Connection) error {
	packet := common.NewPacket(0)

	packet.WriteShort(28)
	packet.WriteString("")
	packet.WriteInt(1)
	packet.WriteInt(2)
	packet.WriteByte(8)

	err := client.Write(packet)

	return err
}
