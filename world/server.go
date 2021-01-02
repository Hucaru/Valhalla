package world

import (
	"log"

	"github.com/Hucaru/Valhalla/common/mnet"
	"github.com/Hucaru/Valhalla/common/mpacket"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
)

// Server data
type Server struct {
	info             internal.World
	login            mnet.Server
	nextPartyID      int32
	reusablePartyIDs []int32
}

// RegisterWithLogin server
func (server *Server) RegisterWithLogin(conn mnet.Server, message string, ribbon byte) {
	server.info.Message = message
	server.info.Ribbon = ribbon

	server.login = conn
	server.registerWithLogin()
}

func (server *Server) registerWithLogin() {
	p := mpacket.CreateInternal(opcode.WorldNew)
	p.WriteString(server.info.Name)
	server.login.Send(p)
}

// ServerDisconnected handler
func (server *Server) ServerDisconnected(conn mnet.Server) {
	for i, v := range server.info.Channels {
		if v.Conn == conn {
			server.info.Channels[i].Conn = nil
			server.info.Channels[i].MaxPop = 0
			server.info.Channels[i].Pop = 0
			server.info.Channels[i].Port = 0
			log.Println("Lost channel", i)
			server.sendChannelInfo()
			break
		}
	}

	server.login.Send(server.info.GenerateInfoPacket())
}

func (server Server) channelBroadcast(p mpacket.Packet) {
	for _, v := range server.info.Channels {
		if v.Conn != nil {
			v.Conn.Send(p)
		}
	}
}

func (server Server) forwardPacketToChannels(conn mnet.Server, reader mpacket.Reader) {
	p := mpacket.NewPacket()
	p.WriteByte(0)
	p.WriteBytes(reader.GetBuffer())
	server.channelBroadcast(p)
}
