package nx

type Item struct {
	price   uint32
	slotMax uint16
}

var Items = make(map[uint32]Item)

func getItemInfo() {

}
