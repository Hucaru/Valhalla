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

	SendCashShopLoadLockerDone      byte = 31
	SendCashShopLoadLockerFailed    byte = 32
	SendCashShopLoadWishDone        byte = 33
	SendCashShopLoadWishFailed      byte = 34
	SendCashShopUpdateWishDone      byte = 35
	SendCashShopUpdateWishFailed    byte = 36
	SendCashShopBuyDone             byte = 37
	SendCashShopBuyFailed           byte = 38
	SendCashShopGiftDone            byte = 39
	SendCashShopGiftFailed          byte = 41
	SendCashShopUseCouponDone       byte = 42
	SendCashShopUseCouponFailed     byte = 44
	SendCashShopUseGiftCouponDone   byte = 45
	SendCashShopIncSlotCountDone    byte = 46
	SendCashShopIncSlotCountFailed  byte = 47
	SendCashShopIncTrunkCountDone   byte = 48
	SendCashShopIncTrunkCountFailed byte = 49

	SendCashShopMoveLtoSDone   byte = 50
	SendCashShopMoveLtoSFailed byte = 51
	SendCashShopMoveStoLDone   byte = 52
	SendCashShopMoveStoLFailed byte = 53
)
