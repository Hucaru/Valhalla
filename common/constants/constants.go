package constants

const (
	MAPLE_VERSION           = 28
	CLIENT_HEADER_SIZE      = 4
	INTERSERVER_HEADER_SIZE = 4
	OPCODE_LENGTH           = 1

	// Opcodes Server -> Client
	LOGIN_RESPONCE        = 0x01
	LOGIN_WORLD_META      = 0x03
	LOGIN_SEND_WORLD_LIST = 0x09
	LOGIN_CHARACTER_DATA  = 0x0A

	// Opcodes Client -> Server
	LOGIN_REQUEST          = 0x01
	LOGIN_CHECK_LOGIN      = 0x08 // wtf is this for?
	LOGIN_CREATE_CHARACTER = 0x09
	LOGIN_WORLD_SELECT     = 0x05
	LOGIN_CHANNEL_SELECT   = 0x04
)
