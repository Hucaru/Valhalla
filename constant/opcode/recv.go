package opcodes

import "github.com/Hucaru/Valhalla/mpacket"

const (
	RecvLoginRequest               mpacket.Opcode = 0x01
	RecvLoginChannelSelect         mpacket.Opcode = 0x04
	RecvLoginWorldSelect           mpacket.Opcode = 0x05
	RecvLoginCheckLogin            mpacket.Opcode = 0x08
	RecvLoginCreateCharacter       mpacket.Opcode = 0x09
	RecvLoginSelectCharacter       mpacket.Opcode = 0x0B
	RecvChannelPlayerLoad          mpacket.Opcode = 0x0C
	RecvLoginNameCheck             mpacket.Opcode = 0x0D
	RecvLoginNewCharacter          mpacket.Opcode = 0x0E
	RecvLoginDeleteChar            mpacket.Opcode = 0x0F
	RecvPing                       mpacket.Opcode = 0x12
	RecvReturnToLoginScreen        mpacket.Opcode = 0x14
	RecvChannelUserPortal          mpacket.Opcode = 0x17
	RecvChannelEnterCashShop       mpacket.Opcode = 0x19
	RecvChannelPlayerMovement      mpacket.Opcode = 0x1A
	RecvChannelPlayerStand         mpacket.Opcode = 0x1B
	RecvChannelPlayerUserChair     mpacket.Opcode = 0x1C
	RecvChannelMeleeSkill          mpacket.Opcode = 0x1D
	RecvChannelRangedSkill         mpacket.Opcode = 0x1E
	RecvChannelMagicSkill          mpacket.Opcode = 0x1F
	RecvChannelDmgRecv             mpacket.Opcode = 0x21
	RecvChannelPlayerSendAllChat   mpacket.Opcode = 0x22
	RecvChannelEmote               mpacket.Opcode = 0x23
	RecvChannelNpcDialogue         mpacket.Opcode = 0x27
	RecvChannelNpcDialogueContinue mpacket.Opcode = 0x28
	RecvChannelNpcShop             mpacket.Opcode = 0x29
	RecvChannelInvMoveItem         mpacket.Opcode = 0x2D
	RecvChannelAddStatPoint        mpacket.Opcode = 0x36
	RecvChannelPassiveRegen        mpacket.Opcode = 0x37
	RecvChannelAddSkillPoint       mpacket.Opcode = 0x38
	RecvChannelSpecialSkill        mpacket.Opcode = 0x39
	RecvChannelCharacterInfo       mpacket.Opcode = 0x3F
	RecvChannelLieDetectorResult   mpacket.Opcode = 0x45
	RecvChannelCharacterReport     mpacket.Opcode = 0x49
	RecvChannelSlashCommands       mpacket.Opcode = 0x4C
	RecvChannelCharacterUIWindow   mpacket.Opcode = 0x4E
	RecvChannelPartyInfo           mpacket.Opcode = 0x4F
	RecvChannelGuildManagement     mpacket.Opcode = 0x51
	RecvChannelGuildReject         mpacket.Opcode = 0x52
	RecvChannelAddBuddy            mpacket.Opcode = 0x55
	RecvChannelMobControl          mpacket.Opcode = 0x6A
	RecvChannelNpcMovement         mpacket.Opcode = 0x6F
)
