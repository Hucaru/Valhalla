package opcode

const (
	RecvCashShopBuyItem        byte = 0x02
	RecvCashShopGiftItem       byte = 0x03
	RecvCashShopUpdateWishlist byte = 0x04
	RecvCashShopIncreaseSlots  byte = 0x05
	RecvCashShopMoveLtoS       byte = 0x0A
	RecvCashShopMoveStoL       byte = 0x0B
	RecvCashShopBuyCoupleRing  byte = 0x18
	RecvCashShopBuyPackage     byte = 0x19
	RecvCashShopGiftPackage    byte = 0x1A
	RecvCashShopBuyQuestItem   byte = 0x1B

	SendCashShopLoadLockerDone      byte = 0x1C
	SendCashShopLoadLockerFailed    byte = 0x1D
	SendCashShopLoadWishDone        byte = 0x1E
	SendCashShopLoadWishFailed      byte = 0x1F
	SendCashShopUpdateWishDone      byte = 0x20
	SendCashShopUpdateWishFailed    byte = 0x21
	SendCashShopBuyDone             byte = 0x22
	SendCashShopBuyFailed           byte = 0x23
	SendCashShopUseCouponDone       byte = 0x24
	SendCashShopUseGiftCouponDone   byte = 0x26
	SendCashShopUseCouponFailed     byte = 0x27
	SendCashShopGiftDone            byte = 0x29
	SendCashShopGiftFailed          byte = 0x2A
	SendCashShopIncSlotCountDone    byte = 0x2B
	SendCashShopIncSlotCountFailed  byte = 0x2C
	SendCashShopIncTrunkCountDone   byte = 0x2D
	SendCashShopIncTrunkCountFailed byte = 0x2E
	SendCashShopMoveLtoSDone        byte = 0x2F
	SendCashShopMoveLtoSFailed      byte = 0x30
	SendCashShopMoveStoLDone        byte = 0x31
	SendCashShopMoveStoLFailed      byte = 0x32
	SendCashShopDeleteDone          byte = 0x33
	SendCashShopDeleteFailed        byte = 0x34
	SendCashShopExpiredDone         byte = 0x35
	SendCashShopExpiredFailed       byte = 0x36
)
