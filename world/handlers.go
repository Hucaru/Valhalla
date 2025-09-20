package world

import (
	"log"
	"math"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// HandleServerPacket from servers
func (server *Server) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.WorldRequestOk:
		server.handleRequestOk(conn, reader)
	case opcode.WorldRequestBad:
		server.handleRequestBad(conn, reader)
	case opcode.ChannelNew:
		server.handleNewChannel(conn, reader)
	case opcode.ChannelInfo:
		server.handleChannelUpdate(conn, reader)
	case opcode.ChannelPlayerConnect:
		server.handlePlayerConnect(conn, reader)
	case opcode.ChannePlayerDisconnect:
		server.handlePlayerDisconnect(conn, reader)
	case opcode.ChannelPlayerBuddyEvent:
		fallthrough
	case opcode.ChannelPlayerChatEvent:
		server.forwardPacketToChannels(conn, reader)
	case opcode.ChannelPlayerPartyEvent:
		server.handlePartyEvent(conn, reader)
	case opcode.ChannelPlayerGuildEvent:
		server.handleGuildEvent(conn, reader)
	case opcode.ChangeRate:
		server.handleChangeRate(conn, reader)
	case opcode.CashShopNew:
		server.handleNewCashShop(conn, reader)
	case opcode.ChannelPlayerMessengerEvent:
		server.handleMessengerEvent(conn, reader)
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

func (server *Server) handleRequestOk(conn mnet.Server, reader mpacket.Reader) {
	server.Info.Name = reader.ReadString(reader.ReadInt16())
	log.Println("Registered as", server.Info.Name, "with login server at", conn)
	server.login.Send(server.Info.GenerateInfoPacket())
}

func (server *Server) handleRequestBad(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Rejected by login server at", conn)
	timer := time.NewTimer(30 * time.Second)

	<-timer.C

	server.registerWithLogin()
}

func (server *Server) handleNewChannel(conn mnet.Server, reader mpacket.Reader) {
	log.Println("New channel request")
	ip := reader.ReadBytes(4)
	port := reader.ReadInt16()
	maxPop := reader.ReadInt16()

	if len(server.Info.Channels) > 19 {
		p := mpacket.CreateInternal(opcode.ChannelBad)
		conn.Send(p)
		return
	}

	pSend := func(id int) {
		p := mpacket.CreateInternal(opcode.ChannelOk)
		p.WriteString(server.Info.Name)
		p.WriteByte(byte(id))
		// Sending the registered channel the world's rates
		p.WriteFloat32(server.Info.Rates.Exp)
		p.WriteFloat32(server.Info.Rates.Drop)
		p.WriteFloat32(server.Info.Rates.Mesos)

		conn.Send(p)
		server.login.Send(server.Info.GenerateInfoPacket())
	}

	// check to see if we have lost any channels
	for i, v := range server.Info.Channels {
		if v.Conn == nil {
			server.Info.Channels[i].Conn = conn
			server.Info.Channels[i].IP = ip
			server.Info.Channels[i].Port = port
			server.Info.Channels[i].MaxPop = maxPop
			pSend(i)

			log.Println("Re-registered channel", i)
			server.sendChannelInfo()
			return
		}
	}
	if server.Info.CashShop.Conn != nil {
		// Send new channel CS info
		p := mpacket.CreateInternal(opcode.CashShopInfo)
		p.WriteBytes(server.Info.CashShop.IP)
		p.WriteInt16(server.Info.CashShop.Port)
		conn.Send(p)
	}

	// TODO highest value party id and set the to current party id if it is larger

	newChannel := internal.Channel{Conn: conn, IP: ip, Port: port, MaxPop: maxPop, Pop: 0}
	server.Info.Channels = append(server.Info.Channels, newChannel)

	pSend(len(server.Info.Channels) - 1)

	log.Println("Registered channel", len(server.Info.Channels)-1)
	server.sendChannelInfo()
}

func (server Server) sendChannelInfo() {
	p := mpacket.CreateInternal(opcode.ChannelConnectionInfo)
	p.WriteByte(byte(len(server.Info.Channels)))

	for _, v := range server.Info.Channels {
		p.WriteBytes(v.IP)
		p.WriteInt16(v.Port)
	}

	server.channelBroadcast(p)
	if server.Info.CashShop.Conn != nil {
		server.Info.CashShop.Conn.Send(p)
	}
}

func (server *Server) handleChannelUpdate(conn mnet.Server, reader mpacket.Reader) {
	id := reader.ReadByte()
	op := reader.ReadByte()
	switch op {
	case 0: //population
		server.Info.Channels[id].Pop = reader.ReadInt16()
	default:
		log.Println("Unkown channel update type", op)
	}
	server.login.Send(server.Info.GenerateInfoPacket())
}

func (server *Server) handleNewCashShop(conn mnet.Server, reader mpacket.Reader) {
	log.Println("New cashshop request")
	ip := reader.ReadBytes(4)
	port := reader.ReadInt16()

	newCashShop := internal.CashShop{Conn: conn, IP: ip, Port: port}
	server.Info.CashShop = newCashShop

	p := mpacket.CreateInternal(opcode.CashShopOk)
	p.WriteString(server.Info.Name)
	conn.Send(p)

	log.Println("Registered CashShop")

	go server.sendCashShopInfo()
}

func (server *Server) sendCashShopInfo() {
	for len(server.Info.Channels) <= 0 {
		log.Println("No channels to send cash shop info to")
		time.Sleep(10 * time.Second)
	}

	p := mpacket.CreateInternal(opcode.CashShopInfo)
	p.WriteBytes(server.Info.CashShop.IP)
	p.WriteInt16(server.Info.CashShop.Port)

	server.channelBroadcast(p)
}

func (server *Server) handlePlayerConnect(conn mnet.Server, reader mpacket.Reader) {
	server.forwardPacketToChannels(conn, reader)

	playerID := reader.ReadInt32()
	_ = reader.ReadString(reader.ReadInt16()) // name
	channelID := reader.ReadByte()            // this needs to be -1 to show player in cash shop
	_ = reader.ReadBool()                     // channelChange
	mapID := reader.ReadInt32()
	_ = reader.ReadInt32() // guildID

	for _, party := range server.parties {
		end := false
		for i, v := range party.PlayerID {
			if v == playerID {
				party.ChannelID[i] = int32(channelID)
				party.MapID[i] = mapID
				server.channelBroadcast(internal.PacketWorldPartyUpdate(party.ID, party.PlayerID[i], int32(i), true, party))
				end = true
				break
			}
		}

		if end {
			break
		}
	}
}

func (server *Server) handlePlayerDisconnect(conn mnet.Server, reader mpacket.Reader) {
	server.forwardPacketToChannels(conn, reader)

	playerID := reader.ReadInt32()
	_ = reader.ReadString(reader.ReadInt16()) // name
	_ = reader.ReadInt32()                    // guildID

	for _, party := range server.parties {
		end := false
		for i, v := range party.PlayerID {
			if v == playerID {
				party.ChannelID[i] = -2 // means offline

				server.channelBroadcast(internal.PacketWorldPartyUpdate(party.ID, party.PlayerID[i], int32(i), true, party))
				end = true
				break
			}
		}

		if end {
			break
		}
	}
}

func (server *Server) handlePartyEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case internal.OpPartyCreate:
		playerID := reader.ReadInt32()
		channelID := reader.ReadByte()
		mapID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		var partyID int32
		if len(server.reusablePartyIDs) > 0 {
			partyID = server.reusablePartyIDs[0]
			server.reusablePartyIDs[0] = server.reusablePartyIDs[len(server.reusablePartyIDs)-1]
			server.reusablePartyIDs = server.reusablePartyIDs[:len(server.reusablePartyIDs)-1]
		} else {
			server.nextPartyID++

			if server.nextPartyID == math.MaxInt32 {
				server.nextPartyID = 1
			}

			partyID = server.nextPartyID
		}

		if partyID == 0 {
			server.Info.Channels[channelID].Conn.Send(internal.PacketWorldPartyCreateApproved(playerID, false, server.parties[partyID]))
		} else {
			server.parties[partyID] = &internal.Party{
				ID: partyID,
			}

			server.parties[partyID].ChannelID[0] = int32(channelID)
			server.parties[partyID].PlayerID[0] = playerID
			server.parties[partyID].Name[0] = name
			server.parties[partyID].MapID[0] = mapID
			server.parties[partyID].Job[0] = job
			server.parties[partyID].Level[0] = level

			server.channelBroadcast(internal.PacketWorldPartyCreateApproved(playerID, true, server.parties[partyID]))
		}

	case internal.OpPartyLeaveExpel:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		kicked := reader.ReadBool()

		if party, ok := server.parties[partyID]; ok {
			for i, v := range party.PlayerID {
				if v == playerID {
					destory := i == 0

					party.ChannelID[i] = 0
					party.PlayerID[i] = 0
					party.Name[i] = ""
					party.MapID[i] = 0
					party.Job[i] = 0
					party.Level[i] = 0

					server.channelBroadcast(internal.PacketWorldPartyLeave(partyID, destory, kicked, int32(i), party))

					server.reusablePartyIDs = append(server.reusablePartyIDs, partyID)
					break
				}
			}
		}
	case internal.OpPartyAccept:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		channelID := reader.ReadInt32()
		mapID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		if party, ok := server.parties[partyID]; ok {
			for i, v := range party.PlayerID {
				if v == 0 { // empty slot
					party.PlayerID[i] = playerID
					party.ChannelID[i] = channelID
					party.MapID[i] = mapID
					party.Job[i] = job
					party.Level[i] = level
					party.Name[i] = name

					server.channelBroadcast(internal.PacketWorldPartyAccept(partyID, playerID, int32(i), party))

					break
				}
			}
		}
	case internal.OpPartyInfoUpdate:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		mapID := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		if party, ok := server.parties[partyID]; ok {
			for i, v := range party.PlayerID {
				if v == playerID {
					party.Job[i] = job
					party.Level[i] = level
					party.Name[i] = name
					party.MapID[i] = mapID

					server.channelBroadcast(internal.PacketWorldPartyUpdate(partyID, playerID, int32(i), false, party))

					break
				}
			}
		}
	default:
		log.Println("Unknown party event type:", op)
	}
}

func (server *Server) handleChangeRate(conn mnet.Server, reader mpacket.Reader) {
	mode := reader.ReadByte()
	rate := reader.ReadFloat32()

	switch mode {
	case 1:
		server.Info.Rates.Exp = rate
	case 2:
		server.Info.Rates.Drop = rate
	case 3:
		server.Info.Rates.Mesos = rate
	}

	if server.Info.Rates.Exp != server.Info.DefaultRates.Exp ||
		server.Info.Rates.Drop != server.Info.DefaultRates.Drop ||
		server.Info.Rates.Mesos != server.Info.DefaultRates.Mesos { // Rates event
		server.Info.Ribbon = 1
		log.Println("GM triggered rates event")
	} else {
		server.Info.Ribbon = 0
	}
	server.login.Send(server.Info.GenerateInfoPacket())

	p := mpacket.CreateInternal(opcode.ChangeRate)
	p.Append(reader.GetBuffer()[1:])

	server.channelBroadcast(p)
}

func (server *Server) handleGuildEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case internal.OpGuildDisband:
		guildID := reader.ReadInt32()

		if _, err := common.DB.Exec("DELETE FROM guilds WHERE (id=?)", guildID); err != nil {
			log.Println(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}

	case internal.OpGuildRankUpdate:
		reader.Skip(4) // guildID
		playerID := reader.ReadInt32()
		rank := reader.ReadByte()

		query := "UPDATE characters set guildRank=? WHERE id=?"

		if _, err := common.DB.Exec(query, rank, playerID); err != nil {
			log.Println(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}
	case internal.OpGuildAddPlayer:
		server.forwardPacketToChannels(conn, reader)
	case internal.OpGuildPointsUpdate:
		guildID := reader.ReadInt32()
		points := reader.ReadInt32()

		query := "UPDATE guilds set points=? WHERE id=?"

		if _, err := common.DB.Exec(query, points, guildID); err != nil {
			log.Fatal(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}
	case internal.OpGuildTitlesChange:
		guildID := reader.ReadInt32()
		master := reader.ReadString(reader.ReadInt16())
		jrMaster := reader.ReadString(reader.ReadInt16())
		member1 := reader.ReadString(reader.ReadInt16())
		member2 := reader.ReadString(reader.ReadInt16())
		member3 := reader.ReadString(reader.ReadInt16())

		query := "UPDATE guilds set master=?, jrMaster=?, member1=?, member2=?, member3=? WHERE id=?"

		if _, err := common.DB.Exec(query, master, jrMaster, member1, member2, member3, guildID); err != nil {
			log.Fatal(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}
	case internal.OpGuildRemovePlayer:
		reader.Skip(4)
		playerID := reader.ReadInt32()

		query := "UPDATE characters set guildID=?, guildRank=? WHERE id=?"

		if _, err := common.DB.Exec(query, nil, 0, playerID); err != nil {
			log.Fatal(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}
	case internal.OpGuildNoticeChange:
		guildID := reader.ReadInt32()
		notice := reader.ReadString(reader.ReadInt16())

		query := "UPDATE guilds SET notice=? WHERE id=?"

		if _, err := common.DB.Exec(query, notice, guildID); err != nil {
			log.Fatal(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}
	case internal.OpGuildEmblemChange:
		guildID := reader.ReadInt32()
		logoBg := reader.ReadInt16()
		logo := reader.ReadInt16()
		logoBgColour := reader.ReadByte()
		logoColour := reader.ReadByte()

		query := "UPDATE guilds SET logoBg=?,logoBgColour=?,logo=?,logoColour=? WHERE id=?"

		if _, err := common.DB.Exec(query, logoBg, logoBgColour, logo, logoColour, guildID); err != nil {
			log.Fatal(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}
	case internal.OpGuildInvite:
		fallthrough
	case internal.OpGuildInviteReject:
		fallthrough
	case internal.OpGuildInviteAccept:
		server.forwardPacketToChannels(conn, reader)
	default:
		log.Println("Unkown guild event type:", op)
	}
}

type kv struct {
	k byte
	v int32
}

type messengerMember struct {
	charID    int32
	name      string
	channelID byte
	slot      byte
	gender    byte
	skin      byte
	face      int32
	hair      int32
	vis       []kv
	hid       []kv
	cashW     int32
	petAcc    int32
}

type messengerRoom struct {
	id      int32
	members [3]*messengerMember
}

func (m *messengerRoom) count() int {
	c := 0
	for _, mm := range m.members {
		if mm != nil {
			c++
		}
	}
	return c
}
func (m *messengerRoom) firstFreeSlot() (byte, bool) {
	for i := 0; i < len(m.members); i++ {
		if m.members[i] == nil {
			return byte(i), true
		}
	}
	return 0, false
}
func (m *messengerRoom) findMemberSlot(id int32) (byte, bool) {
	for i, mm := range m.members {
		if mm != nil && mm.charID == id {
			return byte(i), true
		}
	}
	return 0, false
}

var messengerRooms = map[int32]*messengerRoom{}

func (server *Server) handleMessengerEvent(conn mnet.Server, reader mpacket.Reader) {
	mode := reader.ReadByte()
	senderID := reader.ReadInt32()
	senderChannel := reader.ReadByte()
	senderName := reader.ReadString(reader.ReadInt16())

	switch mode {
	case 0:
		mID := reader.ReadInt32()

		gender := reader.ReadByte()
		skin := reader.ReadByte()
		face := reader.ReadInt32()
		_ = reader.ReadBool()
		hair := reader.ReadInt32()

		vis := make([]kv, 0, 16)
		for {
			b := reader.ReadByte()
			if int8(b) == -1 {
				break
			}
			vis = append(vis, kv{b, reader.ReadInt32()})
		}
		hid := make([]kv, 0, 16)
		for {
			b := reader.ReadByte()
			if int8(b) == -1 {
				break
			}
			hid = append(hid, kv{b, reader.ReadInt32()})
		}
		cashWeapon := reader.ReadInt32()
		petAccessory := reader.ReadInt32()

		room := messengerRooms[mID]
		if mID <= 0 || room == nil {
			room = &messengerRoom{id: senderID}
			messengerRooms[room.id] = room
		}

		if _, exists := room.findMemberSlot(senderID); exists {
			slot, _ := room.findMemberSlot(senderID)
			server.channelBroadcast(packetWorldMessengerSelfEnter(senderID, slot))
			return
		}

		slot, ok := room.firstFreeSlot()
		if !ok {
			return
		}

		room.members[slot] = &messengerMember{
			charID:    senderID,
			name:      senderName,
			channelID: senderChannel,
			slot:      slot,
			gender:    gender,
			skin:      skin,
			face:      face,
			hair:      hair,
			vis:       vis,
			hid:       hid,
			cashW:     cashWeapon,
			petAcc:    petAccessory,
		}

		server.channelBroadcast(packetWorldMessengerSelfEnter(senderID, slot))

		for _, mm := range room.members {
			if mm == nil || mm.charID == senderID {
				continue
			}
			server.channelBroadcast(packetWorldMessengerEnterInline(senderID, mm.slot, mm))
		}
		for _, mm := range room.members {
			if mm == nil || mm.charID == senderID {
				continue
			}
			server.channelBroadcast(packetWorldMessengerEnterInline(mm.charID, slot, room.members[slot]))
		}

	case 2:
		var found *messengerRoom
		for _, r := range messengerRooms {
			if _, ok := r.findMemberSlot(senderID); ok {
				found = r
				break
			}
		}
		if found == nil {
			return
		}
		slot, _ := found.findMemberSlot(senderID)

		for _, mm := range found.members {
			if mm == nil {
				continue
			}
			server.channelBroadcast(packetWorldMessengerLeave(mm.charID, slot))
		}

		found.members[slot] = nil

		if found.count() == 0 {
			delete(messengerRooms, found.id)
		}

	case 3:
		invitee := reader.ReadString(reader.ReadInt16())

		var found *messengerRoom
		for _, r := range messengerRooms {
			if _, ok := r.findMemberSlot(senderID); ok {
				found = r
				break
			}
		}
		if found == nil {
			server.channelBroadcast(packetWorldMessengerInviteResult(senderID, invitee, false))
			return
		}

		for _, ch := range server.Info.Channels {
			if ch.Conn != nil {
				p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
				p.WriteByte(0x03)
				p.WriteInt32(0)
				p.WriteString(senderName)
				p.WriteInt32(found.id)
				p.WriteBool(true)
				p.WriteString(invitee)
				ch.Conn.Send(p)
			}
		}
		server.channelBroadcast(packetWorldMessengerInviteResult(senderID, invitee, true))

	case 5:
		receiver := reader.ReadString(reader.ReadInt16())
		inviter := reader.ReadString(reader.ReadInt16())
		md := reader.ReadByte()
		_ = inviter
		server.channelBroadcast(packetWorldMessengerBlocked(senderID, receiver, md))

	case 6:
		msg := reader.ReadString(reader.ReadInt16())
		var found *messengerRoom
		for _, r := range messengerRooms {
			if _, ok := r.findMemberSlot(senderID); ok {
				found = r
				break
			}
		}
		if found == nil {
			return
		}
		for _, mm := range found.members {
			if mm == nil || mm.charID == senderID {
				continue
			}
			server.channelBroadcast(packetWorldMessengerChat(mm.charID, msg))
		}

	case 7:
		gender := reader.ReadByte()
		skin := reader.ReadByte()
		face := reader.ReadInt32()
		_ = reader.ReadBool()
		hair := reader.ReadInt32()

		vis := make([]kv, 0, 16)
		for {
			b := reader.ReadByte()
			if int8(b) == -1 {
				break
			}
			vis = append(vis, kv{b, reader.ReadInt32()})
		}
		hid := make([]kv, 0, 16)
		for {
			b := reader.ReadByte()
			if int8(b) == -1 {
				break
			}
			hid = append(hid, kv{b, reader.ReadInt32()})
		}
		cashWeapon := reader.ReadInt32()
		petAccessory := reader.ReadInt32()

		var found *messengerRoom
		for _, r := range messengerRooms {
			if _, ok := r.findMemberSlot(senderID); ok {
				found = r
				break
			}
		}
		if found == nil {
			return
		}
		slot, _ := found.findMemberSlot(senderID)
		mm := found.members[slot]
		if mm != nil {
			mm.gender = gender
			mm.skin = skin
			mm.face = face
			mm.hair = hair
			mm.vis = vis
			mm.hid = hid
			mm.cashW = cashWeapon
			mm.petAcc = petAccessory
			mm.channelID = senderChannel
			mm.name = senderName
		}
		for _, other := range found.members {
			if other == nil || other.charID == senderID {
				continue
			}
			server.channelBroadcast(packetWorldMessengerAvatarInline(other.charID, slot, mm))
		}

	default:
		log.Println("Unknown messenger op (channel->world):", mode)
	}
}

// ===== World->Channel packets for Messenger (inline AvatarLook fields) =====

func packetWorldMessengerSelfEnter(recipientID int32, slot byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x00)
	p.WriteInt32(recipientID)
	p.WriteByte(slot)
	return p
}
func packetWorldMessengerEnterInline(recipientID int32, slot byte, m *messengerMember) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x01)
	p.WriteInt32(recipientID)
	p.WriteByte(slot)
	p.WriteByte(m.gender)
	p.WriteByte(m.skin)
	p.WriteInt32(m.face)
	p.WriteBool(true)
	p.WriteInt32(m.hair)
	for _, e := range m.vis {
		p.WriteByte(e.k)
		p.WriteInt32(e.v)
	}
	p.WriteInt8(-1)
	for _, e := range m.hid {
		p.WriteByte(e.k)
		p.WriteInt32(e.v)
	}
	p.WriteInt8(-1)
	p.WriteInt32(m.cashW)
	p.WriteInt32(m.petAcc)
	p.WriteString(m.name)
	p.WriteByte(m.channelID)
	p.WriteBool(true)
	return p
}
func packetWorldMessengerLeave(recipientID int32, slot byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x02)
	p.WriteInt32(recipientID)
	p.WriteByte(slot)
	return p
}
func packetWorldMessengerInviteResult(senderID int32, recipient string, ok bool) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x04)
	p.WriteInt32(senderID)
	p.WriteString(recipient)
	p.WriteBool(ok)
	return p
}
func packetWorldMessengerBlocked(senderID int32, receiver string, mode byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x05)
	p.WriteInt32(senderID)
	p.WriteString(receiver)
	p.WriteByte(mode)
	return p
}
func packetWorldMessengerChat(recipientID int32, msg string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x06)
	p.WriteInt32(recipientID)
	p.WriteString(msg)
	return p
}
func packetWorldMessengerAvatarInline(recipientID int32, slot byte, m *messengerMember) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x07)
	p.WriteInt32(recipientID)
	p.WriteByte(slot)
	p.WriteByte(m.gender)
	p.WriteByte(m.skin)
	p.WriteInt32(m.face)
	p.WriteBool(true)
	p.WriteInt32(m.hair)
	for _, e := range m.vis {
		p.WriteByte(e.k)
		p.WriteInt32(e.v)
	}
	p.WriteInt8(-1)
	for _, e := range m.hid {
		p.WriteByte(e.k)
		p.WriteInt32(e.v)
	}
	p.WriteInt8(-1)
	p.WriteInt32(m.cashW)
	p.WriteInt32(m.petAcc)
	return p
}
