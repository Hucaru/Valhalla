package opcodes

import "github.com/Hucaru/Valhalla/maplepacket"

const (
	RecvLoginRequest               maplepacket.Opcode = 0x01
	RecvLoginChannelSelect         maplepacket.Opcode = 0x04
	RecvLoginWorldSelect           maplepacket.Opcode = 0x05
	RecvLoginCheckLogin            maplepacket.Opcode = 0x08
	RecvLoginCreateCharacter       maplepacket.Opcode = 0x09
	RecvLoginSelectCharacter       maplepacket.Opcode = 0x0B
	RecvChannelPlayerLoad          maplepacket.Opcode = 0x0C
	RecvLoginNameCheck             maplepacket.Opcode = 0x0D
	RecvLoginNewCharacter          maplepacket.Opcode = 0x0E
	RecvLoginDeleteChar            maplepacket.Opcode = 0x0F
	RecvPing                       maplepacket.Opcode = 0x12
	RecvReturnToLoginScreen        maplepacket.Opcode = 0x14
	RecvChannelUserPortal          maplepacket.Opcode = 0x17
	RecvChannelEnterCashShop       maplepacket.Opcode = 0x19
	RecvChannelPlayerMovement      maplepacket.Opcode = 0x1A
	RecvChannelMeleeSkill          maplepacket.Opcode = 0x1D
	RecvChannelRangedSkill         maplepacket.Opcode = 0x1E
	RecvChannelMagicSkill          maplepacket.Opcode = 0x1F
	RecvChannelDmgRecv             maplepacket.Opcode = 0x21
	RecvChannelPlayerSendAllChat   maplepacket.Opcode = 0x22
	RecvChannelEmoticon            maplepacket.Opcode = 0x23
	RecvChannelNpcDialogue         maplepacket.Opcode = 0x27
	RecvChannelNpcDialogueContinue maplepacket.Opcode = 0x28
	RecvChannelNpcShop             maplepacket.Opcode = 0x29
	RecvChannelInvMoveItem         maplepacket.Opcode = 0x2D
	RecvChannelChangeStat          maplepacket.Opcode = 0x36
	RecvChannelPassiveRegen        maplepacket.Opcode = 0x37
	RecvChannelSkillUpdate         maplepacket.Opcode = 0x38
	RecvChannelSpecialSkill        maplepacket.Opcode = 0x39
	RecvChannelCharacterInfo       maplepacket.Opcode = 0x3F
	RecvChannelLieDetectorResult   maplepacket.Opcode = 0x45
	RecvChannelCharacterReport     maplepacket.Opcode = 0x49
	RecvChannelSlashCommands       maplepacket.Opcode = 0x4C
	RecvChannelCharacterUIWindow   maplepacket.Opcode = 0x4E
	RecvChannelPartyInfo           maplepacket.Opcode = 0x4F
	RecvChannelGuildManagement     maplepacket.Opcode = 0x51
	RecvChannelGuildReject         maplepacket.Opcode = 0x52
	RecvChannelAddBuddy            maplepacket.Opcode = 0x55
	RecvChannelMobControl          maplepacket.Opcode = 0x6A
	RecvChannelNpcMovement         maplepacket.Opcode = 0x6F
)
