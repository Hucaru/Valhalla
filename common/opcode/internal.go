package opcode

const (
	WorldNew                byte = 0x01
	WorldRequestOk          byte = 0x02
	WorldRequestBad         byte = 0x03
	WorldInfo               byte = 0x03
	ChannelNew              byte = 0x04
	ChannelOk               byte = 0x05
	ChannelBad              byte = 0x06
	ChannelInfo             byte = 0x07
	ChannelConnectionInfo   byte = 0x08
	ChannePlayerConnect     byte = 0x09
	ChannePlayerDisconnect  byte = 0x0a
	ChannelPlayerChatEvent  byte = 0x0b
	ChannelPlayerBuddyEvent byte = 0x0c
	ChannelPlayerPartyEvent byte = 0x0d
	ChannelPlayerGuildEvent byte = 0x0e
	ChangeRate              byte = 0x0f
)
