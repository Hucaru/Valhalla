package opcode

import "github.com/Hucaru/Valhalla/mpacket"

const (
	SendLoginResponce               mpacket.Opcode = 0x01
	SendLoginWorldMeta              mpacket.Opcode = 0x03
	SendLoginPinRegister            mpacket.Opcode = 0x07 // Add 1 byte, 1 = register mpacket.Opcode a pin
	SendLoginPinStuff               mpacket.Opcode = 0x08 // Setting mpacket.Opcode pin good
	SendLoginWorldList              mpacket.Opcode = 0x09
	SendLoginRestarter              mpacket.Opcode = 0x15
	SendLoginCharacterData          mpacket.Opcode = 0x0A
	SendLoginCharacterMigrate       mpacket.Opcode = 0x0B
	SendLoginNameCheckResult        mpacket.Opcode = 0x0C
	SendLoginNewCharacterGood       mpacket.Opcode = 0x0D
	SendLoginDeleteCharacter        mpacket.Opcode = 0x0E
	SendChannelInventoryOperation   mpacket.Opcode = 0x18
	SendChannelStatChange           mpacket.Opcode = 0x1A
	SendChannelSkillRecordUpdate    mpacket.Opcode = 0x1D
	SendChannelInfoMessage          mpacket.Opcode = 0x20
	SendChannelLieDetectorTest      mpacket.Opcode = 0x23
	SendChannelAvatarInfoWindow     mpacket.Opcode = 0x2c
	SendChannelPartyInfo            mpacket.Opcode = 0x2D
	SendChannelBroadcastMessage     mpacket.Opcode = 0x32
	SendChannelWarpToMap            mpacket.Opcode = 0x36
	SendChannelPortalClosed         mpacket.Opcode = 0x3A
	SendChannelBubblessChat         mpacket.Opcode = 0x3D
	SendChannelWhisper              mpacket.Opcode = 0x3E
	SendChannelEmployee             mpacket.Opcode = 0x43
	SendChannelQuizQAndA            mpacket.Opcode = 0x44
	SendChannelCharacterEnterField  mpacket.Opcode = 0x4E
	SendChannelCharacterLeaveField  mpacket.Opcode = 0x4F
	SendChannelAllChatMsg           mpacket.Opcode = 0x51
	SendChannelPlayerMovement       mpacket.Opcode = 0x65
	SendChannelPlayerUseMeleeSkill  mpacket.Opcode = 0x66
	SendChannelPlayerUseRangedSkill mpacket.Opcode = 0x67
	SendChannelPlayerUseMagicSkill  mpacket.Opcode = 0x68
	SendChannelPlayerTakeDmg        mpacket.Opcode = 0x6B
	SendChannelPlayerEmoticon       mpacket.Opcode = 0x6C
	SendChannelPlayerChangeAvatar   mpacket.Opcode = 0x6F
	SendChannelPlayerAnimation      mpacket.Opcode = 0x70
	SendChannelLevelUpAnimation     mpacket.Opcode = 0x79
	SendChannelShowMob              mpacket.Opcode = 0x86
	SendChannelRemoveMob            mpacket.Opcode = 0x87
	SendChannelControlMob           mpacket.Opcode = 0x88
	SendChannelMoveMob              mpacket.Opcode = 0x8A
	SendChannelControlMobAck        mpacket.Opcode = 0x8B
	SendChannelMobChangeHP          mpacket.Opcode = 0x91
	SendChannelNpcShow              mpacket.Opcode = 0x97
	SendChannelNpcRemove            mpacket.Opcode = 0x98
	SendChannelNpcControl           mpacket.Opcode = 0x99
	SendChannelNpcMovement          mpacket.Opcode = 0x9B
	SendChannelSpawnDoor            mpacket.Opcode = 0xB1
	SendChannelRemoveDoor           mpacket.Opcode = 0xB2
	SendChannelNpcDialogueBox       mpacket.Opcode = 0xC5
	SendChannelNpcShop              mpacket.Opcode = 0xC8
	SendChannelNpcShopResult        mpacket.Opcode = 0xC9
	SendChannelNpcStorage           mpacket.Opcode = 0xCD
	SendChannelRoom                 mpacket.Opcode = 0xDC
	SendChannelRoomBox              mpacket.Opcode = 0x52
)
