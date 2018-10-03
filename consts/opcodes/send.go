package opcodes

var Send send

type send struct {
	LoginResponce                 byte
	LoginWorldMeta                byte
	LoginPinRegister              byte
	LoginPinStuff                 byte
	LoginWorldList                byte
	LoginRestarter                byte
	LoginCharacterData            byte
	LoginCharacterMigrate         byte
	LoginNameCheckResult          byte
	LoginNewCharacterGood         byte
	LoginDeleteCharacter          byte
	ChannelInventoryOperation     byte
	ChannelStatChange             byte
	ChannelSkillRecordUpdate      byte
	ChannelInfoMessage            byte
	ChannelLieDetectorTest        byte
	ChannelAvatarInfoWindow       byte
	ChannelPartyInfo              byte
	ChannelBroadcastMessage       byte
	ChannelWarpToMap              byte
	ChannelPortalClosed           byte
	ChannelBubblessChat           byte
	ChannelWhisper                byte
	ChannelEmployee               byte
	ChannelQuizQAndA              byte
	ChannelCharacterEnterField    byte
	ChannelCharacterLeaveField    byte
	ChannelAllChatMsg             byte
	ChannelPlayerMovement         byte
	ChannelPlayerUseStandardSkill byte
	ChannelPlayerUseRangedSkill   byte
	ChannelPlayerUseMagicSkill    byte
	ChannelPlayerTakeDmg          byte
	ChannelPlayerEmoticon         byte
	ChannelPlayerChangeAvatar     byte
	ChannelPlayerAnimation        byte
	ChannelLevelUpAnimation       byte
	ChannelShowMob                byte
	ChannelRemoveMob              byte
	ChannelControlMob             byte
	ChannelMoveMob                byte
	ChannelControlMobAck          byte
	ChannelMobChangeHP            byte
	ChannelNpcShow                byte
	ChannelNpcRemove              byte
	ChannelNpcControl             byte
	ChannelNpcMovement            byte
	ChannelSpawnDoor              byte
	ChannelRemoveDoor             byte
	ChannelNpcDialogueBox         byte
	ChannelNpcShop                byte
	ChannelNpcShopResult          byte
	ChannelNpcStorage             byte
	ChannelRoom                   byte
	ChannelRoomBox                byte
}

func init() {
	Send.LoginResponce = 0x01
	Send.LoginWorldMeta = 0x03
	Send.LoginPinRegister = 0x07 // Add 1 byte, 1 = register a pin
	Send.LoginPinStuff = 0x08    // Setting pin good
	Send.LoginWorldList = 0x09
	Send.LoginRestarter = 0x15
	Send.LoginCharacterData = 0x0A
	Send.LoginCharacterMigrate = 0x0B
	Send.LoginNameCheckResult = 0x0C
	Send.LoginNewCharacterGood = 0x0D
	Send.LoginDeleteCharacter = 0x0E
	Send.ChannelInventoryOperation = 0x18
	Send.ChannelStatChange = 0x1A
	Send.ChannelSkillRecordUpdate = 0x1D
	Send.ChannelInfoMessage = 0x20
	Send.ChannelLieDetectorTest = 0x23
	Send.ChannelAvatarInfoWindow = 0x2c
	Send.ChannelPartyInfo = 0x2D
	Send.ChannelBroadcastMessage = 0x32
	Send.ChannelWarpToMap = 0x36
	Send.ChannelPortalClosed = 0x3A
	Send.ChannelBubblessChat = 0x3D
	Send.ChannelWhisper = 0x3E
	Send.ChannelEmployee = 0x43
	Send.ChannelQuizQAndA = 0x44
	Send.ChannelCharacterEnterField = 0x4E
	Send.ChannelCharacterLeaveField = 0x4F
	Send.ChannelAllChatMsg = 0x51
	Send.ChannelPlayerMovement = 0x65
	Send.ChannelPlayerUseStandardSkill = 0x66
	Send.ChannelPlayerUseRangedSkill = 0x67
	Send.ChannelPlayerUseMagicSkill = 0x68
	Send.ChannelPlayerTakeDmg = 0x6B
	Send.ChannelPlayerEmoticon = 0x6C
	Send.ChannelPlayerChangeAvatar = 0x6F
	Send.ChannelPlayerAnimation = 0x70
	Send.ChannelLevelUpAnimation = 0x79
	Send.ChannelShowMob = 0x86
	Send.ChannelRemoveMob = 0x87
	Send.ChannelControlMob = 0x88
	Send.ChannelMoveMob = 0x8A
	Send.ChannelControlMobAck = 0x8B
	Send.ChannelMobChangeHP = 0x91
	Send.ChannelNpcShow = 0x97
	Send.ChannelNpcRemove = 0x98
	Send.ChannelNpcControl = 0x99
	Send.ChannelNpcMovement = 0x9B
	Send.ChannelSpawnDoor = 0xB1
	Send.ChannelRemoveDoor = 0xB2
	Send.ChannelNpcDialogueBox = 0xC5
	Send.ChannelNpcShop = 0xC8
	Send.ChannelNpcShopResult = 0xC9
	Send.ChannelNpcStorage = 0xCD
	Send.ChannelRoom = 0xDC
	Send.ChannelRoomBox = 0x52
}
