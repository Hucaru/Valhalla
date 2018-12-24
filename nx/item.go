package nx

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hucaru/gonx"
)

// Item data from nx
type Item struct {
	InvTabID                                                       byte
	Cash                                                           bool
	Only, TradeBlock, ExpireOnLogout, Quest, TimeLimited           int64
	ReqLevel                                                       byte
	Tuc                                                            byte // Total upgrade count?
	SlotMax                                                        int16
	ReqJob                                                         int64
	ReqSTR, ReqDEX, ReqINT, ReqLUK, IncSTR, IncDEX, IncINT, IncLUK int16
	IncACC, IncEVA, IncMDD, IncPDD, IncMAD, IncPAD, IncMHP, IncMMP int16
	AttackSpeed, Attack, IncJump, IncSpeed, RecoveryHP             int64
	Price                                                          int32
	NotSale                                                        int64
	UnitPrice                                                      float64
	Life, Hungry                                                   int64
	PickupItem, PickupAll, SweepForDrop                            int64
	ConsumeHP, LongRange                                           int64
	Recovery                                                       float64
	ReqPOP                                                         int64 // ?
	NameTag                                                        int64
	Pachinko                                                       int64
	VSlot, ISlot                                                   string
	Type                                                           int64
	Success                                                        int64 // Scroll type
	Cursed                                                         int64
	Add                                                            int64 // ?
	DropSweep                                                      int64
	Rate                                                           int64
	Meso                                                           int64
	Path                                                           string
	FloatType                                                      int64
	NoFlip                                                         string
	StateChangeItem                                                int64
	BigSize                                                        int64
	Sfx                                                            string
	Walk                                                           int64
	AfterImage                                                     string
	Stand                                                          int64
	Knockback                                                      int64
	Fs                                                             int64
	ChatBalloon                                                    int64
}

func extractItems(nodes []gonx.Node, textLookup []string) map[int32]Item {
	items := make(map[int32]Item)

	searches := []string{"/Character/Accessory", "/Character/Cap", "/Character/Cape", "/Character/Coat",
		"/Character/Face", "/Character/Glove", "/Character/Hair", "/Character/Longcoat", "/Character/Pants",
		"/Character/PetEquip", "/Character/Ring", "/Character/Shield", "/Character/Shoes", "/Character/Weapon",
		"Item/Pet"}

	for _, search := range searches {
		valid := gonx.FindNode(search, nodes, textLookup, func(node *gonx.Node) {
			for i := uint32(0); i < uint32(node.ChildCount); i++ {
				itemNode := nodes[node.ChildID+i]
				name := textLookup[itemNode.NameID]
				subSearch := search + "/" + name + "/info"

				var item Item

				valid := gonx.FindNode(subSearch, nodes, textLookup, func(node *gonx.Node) {
					item = getItem(node, nodes, textLookup)
				})

				if !valid {
					log.Println("Invalid node search:", subSearch)
				}

				name = strings.TrimSuffix(name, filepath.Ext(name))
				itemID, err := strconv.Atoi(name)

				if err != nil {
					log.Println(err)
					continue
				}

				item.InvTabID = byte(itemID / 1e6)
				items[int32(itemID)] = item
			}
		})

		if !valid {
			log.Println("Invalid node search:", search)
		}
	}

	searches = []string{"/Item/Cash", "/Item/Consume", "/Item/Etc", "/Item/Install"}

	for _, search := range searches {
		valid := gonx.FindNode(search, nodes, textLookup, func(node *gonx.Node) {
			for i := uint32(0); i < uint32(node.ChildCount); i++ {
				itemGroupNode := nodes[node.ChildID+i]
				groupName := textLookup[itemGroupNode.NameID]

				for j := uint32(0); j < uint32(itemGroupNode.ChildCount); j++ {
					itemNode := nodes[itemGroupNode.ChildID+j]
					name := textLookup[itemNode.NameID]

					subSearch := search + "/" + groupName + "/" + name + "/info"

					var item Item

					valid := gonx.FindNode(subSearch, nodes, textLookup, func(node *gonx.Node) {
						item = getItem(node, nodes, textLookup)
					})

					if !valid {
						log.Println("Invalid node search:", subSearch)
					}

					name = strings.TrimSuffix(name, filepath.Ext(name))
					itemID, err := strconv.Atoi(name)

					if err != nil {
						log.Println(err)
						continue
					}

					item.InvTabID = byte(itemID / 1e6)
					items[int32(itemID)] = item
				}
			}
		})

		if !valid {
			log.Println("Invalid node search:", search)
		}
	}

	return items
}

func getItem(node *gonx.Node, nodes []gonx.Node, textLookup []string) Item {
	item := Item{}

	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		option := nodes[node.ChildID+i]
		optionName := textLookup[option.NameID]

		switch optionName {
		case "cash":
			item.Cash = gonx.DataToBool(option.Data[0])
		case "reqSTR":
			item.ReqSTR = gonx.DataToInt16(option.Data)
		case "reqDEX":
			item.ReqDEX = gonx.DataToInt16(option.Data)
		case "reqINT":
			item.ReqINT = gonx.DataToInt16(option.Data)
		case "reqLUK":
			item.ReqLUK = gonx.DataToInt16(option.Data)
		case "reqJob":
			item.ReqJob = gonx.DataToInt64(option.Data)
		case "reqLevel":
			item.ReqLevel = option.Data[0]
		case "price":
			item.Price = gonx.DataToInt32(option.Data)
		case "incSTR":
			item.IncSTR = gonx.DataToInt16(option.Data)
		case "incDEX":
			item.IncDEX = gonx.DataToInt16(option.Data)
		case "incINT":
			item.IncINT = gonx.DataToInt16(option.Data)
		case "incLUK": // typo?
			fallthrough
		case "incLUk":
			item.IncLUK = gonx.DataToInt16(option.Data)
		case "incMMD": // typo?
			fallthrough
		case "incMDD":
			item.IncMDD = gonx.DataToInt16(option.Data)
		case "incPDD":
			item.IncPDD = gonx.DataToInt16(option.Data)
		case "incMAD":
			item.IncMAD = gonx.DataToInt16(option.Data)
		case "incPAD":
			item.IncPAD = gonx.DataToInt16(option.Data)
		case "incEVA":
			item.IncEVA = gonx.DataToInt16(option.Data)
		case "incACC":
			item.IncACC = gonx.DataToInt16(option.Data)
		case "incMHP":
			item.IncMHP = gonx.DataToInt16(option.Data)
		case "recoveryHP":
			item.RecoveryHP = gonx.DataToInt64(option.Data)
		case "incMMP":
			item.IncMMP = gonx.DataToInt16(option.Data)
		case "only":
			item.Only = gonx.DataToInt64(option.Data)
		case "attackSpeed":
			item.AttackSpeed = gonx.DataToInt64(option.Data)
		case "attack":
			item.Attack = gonx.DataToInt64(option.Data)
		case "incSpeed":
			item.IncSpeed = gonx.DataToInt64(option.Data)
		case "incJump":
			item.IncJump = gonx.DataToInt64(option.Data)
		case "tuc": // total upgrade count?
			item.Tuc = option.Data[0]
		case "notSale":
			item.NotSale = gonx.DataToInt64(option.Data)
		case "tradeBlock":
			item.TradeBlock = gonx.DataToInt64(option.Data)
		case "expireOnLogout":
			item.ExpireOnLogout = gonx.DataToInt64(option.Data)
		case "slotMax":
			item.SlotMax = gonx.DataToInt16(option.Data)
		case "quest":
			item.Quest = gonx.DataToInt64(option.Data)
		case "life":
			item.Life = gonx.DataToInt64(option.Data)
		case "hungry":
			item.Hungry = gonx.DataToInt64(option.Data)
		case "pickupItem":
			item.PickupItem = gonx.DataToInt64(option.Data)
		case "pickupAll":
			item.PickupAll = gonx.DataToInt64(option.Data)
		case "sweepForDrop":
			item.SweepForDrop = gonx.DataToInt64(option.Data)
		case "longRange":
			item.LongRange = gonx.DataToInt64(option.Data)
		case "consumeHP":
			item.ConsumeHP = gonx.DataToInt64(option.Data)
		case "unitPrice":
			item.UnitPrice = gonx.DataToFloat64(option.Data)
		case "timeLimited":
			item.TimeLimited = gonx.DataToInt64(option.Data)
		case "recovery":
			item.Recovery = gonx.DataToFloat64(option.Data)
		case "regPOP":
			fallthrough
		case "reqPOP":
			item.ReqPOP = gonx.DataToInt64(option.Data)
		case "nameTag":
			item.NameTag = gonx.DataToInt64(option.Data)
		case "pachinko":
			item.Pachinko = gonx.DataToInt64(option.Data)
		case "vslot":
			item.VSlot = textLookup[gonx.DataToUint32(option.Data)]
		case "islot":
			item.ISlot = textLookup[gonx.DataToUint32(option.Data)]
		case "type":
			item.Type = gonx.DataToInt64(option.Data)
		case "success":
			item.Success = gonx.DataToInt64(option.Data)
		case "cursed":
			item.Cursed = gonx.DataToInt64(option.Data)
		case "add":
			item.Add = gonx.DataToInt64(option.Data)
		case "dropSweep":
			item.DropSweep = gonx.DataToInt64(option.Data)
		case "time":
		case "rate":
			item.Rate = gonx.DataToInt64(option.Data)
		case "meso":
			item.Meso = gonx.DataToInt64(option.Data)
		case "path":
			idLookup := gonx.DataToUint32(option.Data)
			item.Path = textLookup[idLookup]
		case "floatType":
			item.FloatType = gonx.DataToInt64(option.Data)
		case "noFlip":
			item.NoFlip = textLookup[gonx.DataToUint32(option.Data)]
		case "stateChangeItem":
			item.StateChangeItem = gonx.DataToInt64(option.Data)
		case "bigSize":
			item.BigSize = gonx.DataToInt64(option.Data)
		case "icon":
		case "iconRaw":
		case "sfx":
			item.Sfx = textLookup[gonx.DataToUint32(option.Data)]
		case "walk":
			item.Walk = gonx.DataToInt64(option.Data)
		case "afterImage":
			item.AfterImage = textLookup[gonx.DataToUint32(option.Data)]
		case "stand":
			item.Stand = gonx.DataToInt64(option.Data)
		case "knockback":
			item.Knockback = gonx.DataToInt64(option.Data)
		case "fs":
			item.Fs = gonx.DataToInt64(option.Data)
		case "chatBalloon":
			item.ChatBalloon = gonx.DataToInt64(option.Data)
		case "sample":
		case "iconD":
		case "iconRawD":
		case "iconReward":
		default:
			log.Println("Unsupported NX item option:", optionName, "->", option.Data)
		}

	}

	return item
}
