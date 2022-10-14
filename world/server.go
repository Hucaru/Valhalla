package world

import (
	"log"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// Server data
type Server struct {
	Info             internal.World
	login            mnet.Server
	nextPartyID      int32
	reusablePartyIDs []int32
}

// RegisterWithLogin server
func (server *Server) RegisterWithLogin(conn mnet.Server) {
	server.login = conn
	server.registerWithLogin()
}

func (server *Server) registerWithLogin() {
	p := mpacket.CreateInternal(opcode.WorldNew)
	p.WriteString(server.Info.Name)
	server.login.Send(p)
}

// ServerDisconnected handler
func (server *Server) ServerDisconnected(conn mnet.Server) {
	for i, v := range server.Info.Channels {
		if v.Conn == conn {
			server.Info.Channels[i].Conn = nil
			server.Info.Channels[i].MaxPop = 0
			server.Info.Channels[i].Pop = 0
			server.Info.Channels[i].Port = 0
			log.Println("Lost channel", i)
			server.sendChannelInfo()
			break
		}
	}

	server.login.Send(server.Info.GenerateInfoPacket())
}

func (server Server) channelBroadcast(p mpacket.Packet) {
	for _, v := range server.Info.Channels {
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
