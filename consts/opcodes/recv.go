package opcodes

var Recv recv

type recv struct {
	LoginRequest               byte
	LoginChannelSelect         byte
	LoginWorldSelect           byte
	LoginCheckLogin            byte
	LoginCreateCharacter       byte
	LoginSelectCharacter       byte
	LoginNameCheck             byte
	LoginNewCharacter          byte
	LoginDeleteChar            byte
	Ping                       byte
	ReturnToLoginScreen        byte
	ChannelPlayerLoad          byte
	ChannelUserPortal          byte
	ChannelEnterCashShop       byte
	ChannelPlayerMovement      byte
	ChannelStandardSkill       byte
	ChannelRangedSkill         byte
	ChannelMagicSkill          byte
	ChannelDmgRecv             byte
	ChannelPlayerSendAllChat   byte
	ChannelEmoticon            byte
	ChannelNpcDialogue         byte
	ChannelNpcDialogueContinue byte
	ChannelNpcShop             byte
	ChannelInvMoveItem         byte
	ChannelChangeStat          byte
	ChannelPassiveRegen        byte
	ChannelSkillUpdate         byte
	ChannelSpecialSkill        byte
	ChannelCharacterInfo       byte
	ChannelLieDetectorResult   byte
	ChannelCharacterReport     byte
	ChannelSlashCommands       byte
	ChannelCharacterUIWindow   byte
	ChannelPartyInfo           byte
	ChannelGuildManagement     byte
	ChannelGuildReject         byte
	ChannelAddBuddy            byte
	ChannelMobControl          byte
	ChannelNpcMovement         byte
}

func init() {
	Recv.LoginRequest = 0x01
	Recv.LoginChannelSelect = 0x04
	Recv.LoginWorldSelect = 0x05
	Recv.LoginCheckLogin = 0x08
	Recv.LoginCreateCharacter = 0x09
	Recv.LoginSelectCharacter = 0x0B
	Recv.LoginNameCheck = 0x0D
	Recv.LoginNewCharacter = 0x0E
	Recv.LoginDeleteChar = 0x0F
	Recv.Ping = 0x12
	Recv.ReturnToLoginScreen = 0x14
	Recv.ChannelPlayerLoad = 0x0C
	Recv.ChannelUserPortal = 0x17
	Recv.ChannelEnterCashShop = 0x19
	Recv.ChannelPlayerMovement = 0x1A
	Recv.ChannelStandardSkill = 0x1D
	Recv.ChannelRangedSkill = 0x1E
	Recv.ChannelMagicSkill = 0x1F
	Recv.ChannelDmgRecv = 0x21
	Recv.ChannelPlayerSendAllChat = 0x22
	Recv.ChannelEmoticon = 0x23
	Recv.ChannelNpcDialogue = 0x27
	Recv.ChannelNpcDialogueContinue = 0x28
	Recv.ChannelNpcShop = 0x29
	Recv.ChannelInvMoveItem = 0x2D
	Recv.ChannelChangeStat = 0x36
	Recv.ChannelPassiveRegen = 0x37
	Recv.ChannelSkillUpdate = 0x38
	Recv.ChannelSpecialSkill = 0x39
	Recv.ChannelCharacterInfo = 0x3F
	Recv.ChannelLieDetectorResult = 0x45
	Recv.ChannelCharacterReport = 0x49
	Recv.ChannelSlashCommands = 0x4C
	Recv.ChannelCharacterUIWindow = 0x4E
	Recv.ChannelPartyInfo = 0x4F
	Recv.ChannelGuildManagement = 0x51
	Recv.ChannelGuildReject = 0x52
	Recv.ChannelAddBuddy = 0x55
	Recv.ChannelMobControl = 0x6A
	Recv.ChannelNpcMovement = 0x6F
}
