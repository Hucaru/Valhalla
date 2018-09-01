package constants

// Generic Constants
const (
	MapleVersion          = 28
	ClientHeaderSize      = 4
	InterserverHeaderSize = 4
	OpcodeLength          = 1
)

// Opcodes
const (
	// Server -> Client
	SendLoginResponce         = 0x01
	SendLoginWorldMeta        = 0x03
	SendLoginPinRegister      = 0x07 // Add 1 byte, 1 = register a pin
	SendLoginPinStuff         = 0x08 // Setting pin good
	SendLoginWorldList        = 0x09
	SendLoginRestarter        = 0x15
	SendLoginCharacterData    = 0x0A
	SendLoginCharacterMigrate = 0x0B
	SendLoginNameCheckResult  = 0x0C
	SendLoginNewCharacterGood = 0x0D
	SendLoginDeleteCharacter  = 0x0E

	SendChannelInventoryOperation     = 0x18
	SendChannelStatChange             = 0x1A
	SendChannelSkillRecordUpdate      = 0x1D
	SendChannelInfoMessage            = 0x20
	SendChannelLieDetectorTest        = 0x23
	SendChannelAvatarInfoWindow       = 0x2c
	SendChannelPartyInfo              = 0x2D
	SendChannelBroadcastMessage       = 0x32
	SendChannelWarpToMap              = 0x36
	SendChannelPortalClosed           = 0x3A
	SendChannelBubblessChat           = 0x3D
	SendChannelWhisper                = 0x3E
	SendChannelEmployee               = 0x43
	SendChannelQuizQAndA              = 0x44
	SendChannelCharacterEnterField    = 0x4E
	SendChannelCharacterLeaveField    = 0x4F
	SendChannelAllChatMsg             = 0x51
	SendChannelPlayerMovement         = 0x65
	SendChannelPlayerUseStandardSkill = 0x66
	SendChannelPlayerUseRangedSkill   = 0x67
	SendChannelPlayerUseMagicSkill    = 0x68
	SendChannelPlayerTakeDmg          = 0x6B
	SendChannelPlayerEmoticon         = 0x6C
	SendChannelPlayerChangeAvatar     = 0x6F
	SendChannelPlayerAnimation        = 0x70
	SendChannelLevelUpAnimation       = 0x79
	SendChannelShowMob                = 0x86
	SendChannelRemoveMob              = 0x87
	SendChannelControlMob             = 0x88
	SendChannelMoveMob                = 0x8A
	SendChannelControlMobAck          = 0x8B
	SendChannelMobChangeHP            = 0x91
	SendChannelNpcShow                = 0x97
	SendChannelNpcRemove              = 0x98
	SendChannelNpcControl             = 0x99
	SendChannelNpcMovement            = 0x9B
	SendChannelSpawnDoor              = 0xB1
	SendChannelRemoveDoor             = 0xB2
	SendChannelNpcDialogueBox         = 0xC5
	SendChannelNpcShop                = 0xC8
	SendChannelNpcShopResult          = 0xC9
	SendChannelNpcStorage             = 0xCD
	SendChannelRoom                   = 0xDC
	SendChannelRoomBox                = 0x52

	// Client -> Server
	RecvLoginRequest         = 0x01
	RecvLoginChannelSelect   = 0x04
	RecvLoginWorldSelect     = 0x05
	RecvLoginCheckLogin      = 0x08
	RecvLoginCreateCharacter = 0x09
	RecvLoginSelectCharacter = 0x0B
	RecvLoginNameCheck       = 0x0D
	RecvLoginNewCharacter    = 0x0E
	RecvLoginDeleteChar      = 0x0F
	RecvPing                 = 0x12
	RecvReturnToLoginScreen  = 0x14

	RecvChannelPlayerLoad          = 0x0C
	RecvChannelUserPortal          = 0x17
	RecvChannelEnterCashShop       = 0x19
	RecvChannelPlayerMovement      = 0x1A
	RecvChannelStandardSkill       = 0x1D
	RecvChannelRangedSkill         = 0x1E
	RecvChannelMagicSkill          = 0x1F
	RecvChannelDmgRecv             = 0x21
	RecvChannelPlayerSendAllChat   = 0x22
	RecvChannelEmoticon            = 0x23
	RecvChannelNpcDialogue         = 0x27
	RecvChannelNpcDialogueContinue = 0x28
	RecvChannelNpcShop             = 0x29
	RecvChannelInvMoveItem         = 0x2D
	RecvChannelChangeStat          = 0x36
	RecvChannelPassiveRegen        = 0x37
	RecvChannelSkillUpdate         = 0x38
	RecvChannelSpecialSkill        = 0x39
	RecvChannelCharacterInfo       = 0x3F
	RecvChannelLieDetectorResult   = 0x45
	RecvChannelCharacterReport     = 0x49
	RecvChannelSlashCommands       = 0x4C
	RecvChannelCharacterUIWindow   = 0x4E
	RecvChannelPartyInfo           = 0x4F
	RecvChannelGuildManagement     = 0x51
	RecvChannelGuildReject         = 0x52
	RecvChannelAddBuddy            = 0x55
	RecvChannelMobControl          = 0x6A
	RecvChannelNpcMovement         = 0x6F
)
