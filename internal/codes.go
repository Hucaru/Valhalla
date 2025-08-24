package internal

const (
	OpChatWhispher = 0x00
	OpChatBuddy    = 0x01
	OpChatParty    = 0x02
	OpChatGuild    = 0x03

	OpPartyCreate     = 0x01
	OpPartyLeaveExpel = 0x02
	OpPartyAccept     = 0x03
	OpPartyInfoUpdate = 0x04

	OpGuildDisband      = 0x01
	OpGuildRankUpdate   = 0x02
	OpGuildAddPlayer    = 0x03
	OpGuildRemovePlayer = 0x04
	OpGuildNoticeChange = 0x05
	OpGuildEmblemChange = 0x06
	OpGuildPointsUpdate = 0x07
	OpGuildTitlesChange = 0x08
	OpGuildInvite       = 0x09
	OpGuildInviteReject = 0x0a
	OpGuildInviteAccept = 0x0b

	OpQuestLostItem  = 0x00
	OpQuestStarted   = 0x01
	OpQuestCompleted = 0x02
	OpQuestForfeit   = 0x03

	OpFameNotifySource  = 0x00
	OpFameIncorrectUser = 0x01
	OpFameUnderLevel    = 0x02
	OpFameThisDay       = 0x03
	OpFameThisMonth     = 0x04
	OpFameNotifyTarget  = 0x05
)
