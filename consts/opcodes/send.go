package opcodes

import "github.com/Hucaru/Valhalla/maplepacket"

const (
	SendLoginResponce               maplepacket.Opcode = 0x01
	SendLoginWorldMeta              maplepacket.Opcode = 0x03
	SendLoginPinRegister            maplepacket.Opcode = 0x07 // Add 1 byte, 1 = register maplepacket.Opcode a pin
	SendLoginPinStuff               maplepacket.Opcode = 0x08 // Setting maplepacket.Opcode pin good
	SendLoginWorldList              maplepacket.Opcode = 0x09
	SendLoginRestarter              maplepacket.Opcode = 0x15
	SendLoginCharacterData          maplepacket.Opcode = 0x0A
	SendLoginCharacterMigrate       maplepacket.Opcode = 0x0B
	SendLoginNameCheckResult        maplepacket.Opcode = 0x0C
	SendLoginNewCharacterGood       maplepacket.Opcode = 0x0D
	SendLoginDeleteCharacter        maplepacket.Opcode = 0x0E
	SendChannelInventoryOperation   maplepacket.Opcode = 0x18
	SendChannelStatChange           maplepacket.Opcode = 0x1A
	SendChannelSkillRecordUpdate    maplepacket.Opcode = 0x1D
	SendChannelInfoMessage          maplepacket.Opcode = 0x20
	SendChannelLieDetectorTest      maplepacket.Opcode = 0x23
	SendChannelAvatarInfoWindow     maplepacket.Opcode = 0x2c
	SendChannelPartyInfo            maplepacket.Opcode = 0x2D
	SendChannelBroadcastMessage     maplepacket.Opcode = 0x32
	SendChannelWarpToMap            maplepacket.Opcode = 0x36
	SendChannelPortalClosed         maplepacket.Opcode = 0x3A
	SendChannelBubblessChat         maplepacket.Opcode = 0x3D
	SendChannelWhisper              maplepacket.Opcode = 0x3E
	SendChannelEmployee             maplepacket.Opcode = 0x43
	SendChannelQuizQAndA            maplepacket.Opcode = 0x44
	SendChannelCharacterEnterField  maplepacket.Opcode = 0x4E
	SendChannelCharacterLeaveField  maplepacket.Opcode = 0x4F
	SendChannelAllChatMsg           maplepacket.Opcode = 0x51
	SendChannelPlayerMovement       maplepacket.Opcode = 0x65
	SendChannelPlayerUseMeleeSkill  maplepacket.Opcode = 0x66
	SendChannelPlayerUseRangedSkill maplepacket.Opcode = 0x67
	SendChannelPlayerUseMagicSkill  maplepacket.Opcode = 0x68
	SendChannelPlayerTakeDmg        maplepacket.Opcode = 0x6B
	SendChannelPlayerEmoticon       maplepacket.Opcode = 0x6C
	SendChannelPlayerChangeAvatar   maplepacket.Opcode = 0x6F
	SendChannelPlayerAnimation      maplepacket.Opcode = 0x70
	SendChannelLevelUpAnimation     maplepacket.Opcode = 0x79
	SendChannelShowMob              maplepacket.Opcode = 0x86
	SendChannelRemoveMob            maplepacket.Opcode = 0x87
	SendChannelControlMob           maplepacket.Opcode = 0x88
	SendChannelMoveMob              maplepacket.Opcode = 0x8A
	SendChannelControlMobAck        maplepacket.Opcode = 0x8B
	SendChannelMobChangeHP          maplepacket.Opcode = 0x91
	SendChannelNpcShow              maplepacket.Opcode = 0x97
	SendChannelNpcRemove            maplepacket.Opcode = 0x98
	SendChannelNpcControl           maplepacket.Opcode = 0x99
	SendChannelNpcMovement          maplepacket.Opcode = 0x9B
	SendChannelSpawnDoor            maplepacket.Opcode = 0xB1
	SendChannelRemoveDoor           maplepacket.Opcode = 0xB2
	SendChannelNpcDialogueBox       maplepacket.Opcode = 0xC5
	SendChannelNpcShop              maplepacket.Opcode = 0xC8
	SendChannelNpcShopResult        maplepacket.Opcode = 0xC9
	SendChannelNpcStorage           maplepacket.Opcode = 0xCD
	SendChannelRoom                 maplepacket.Opcode = 0xDC
	SendChannelRoomBox              maplepacket.Opcode = 0x52
)
