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
	Cash, Pet                                                      bool
	Only, TradeBlock, ExpireOnLogout, Quest, TimeLimited           int64
	ReqLevel                                                       byte
	Tuc                                                            byte // Total upgrade count?
	SlotMax                                                        int16
	ReqJob                                                         int64
	ReqSTR, ReqDEX, ReqINT, ReqLUK, IncSTR, IncDEX, IncINT, IncLUK int16
	IncACC, IncEVA, IncMDD, IncPDD, IncMAD, IncPAD, IncMHP, IncMMP float64
	Attack, IncJump, IncSpeed, RecoveryHP                          float64
	HP, MP                                                         int16
	AttackSpeed                                                    int16
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
	MoveTo                                                         int32
}

func extractItems(nodes []gonx.Node, textLookup []string) map[int32]Item {
	items := make(map[int32]Item)

	// Character-equipment structure: /Character/<Category>/<ItemID>.img/info
	characterSearches := []string{
		"/Character/Accessory", "/Character/Cap", "/Character/Cape", "/Character/Coat",
		"/Character/Face", "/Character/Glove", "/Character/Hair", "/Character/Longcoat", "/Character/Pants",
		"/Character/PetEquip", "/Character/Ring", "/Character/Shield", "/Character/Shoes", "/Character/Weapon",
	}

	for _, base := range characterSearches {
		ok := gonx.FindNode(base, nodes, textLookup, func(node *gonx.Node) {
			iterateChildren(node, nodes, textLookup, func(itemNode gonx.Node, name string) {
				subSearch := base + "/" + name + "/info"
				var itm Item
				if !findAndExtract(subSearch, nodes, textLookup, &itm) {
					log.Println("Invalid node search:", subSearch)
					return
				}
				if !addItemByName(name, &itm, items) {
					return
				}
			})
		})
		if !ok {
			log.Println("Invalid node search:", base)
		}
	}

	// Item groups like Cash/Etc/Install: /Item/<Group>/<SubGroup>/<ItemID>.img/(info|spec)
	groupedSearches := []string{"/Item/Cash", "/Item/Etc", "/Item/Install"}
	for _, base := range groupedSearches {
		ok := gonx.FindNode(base, nodes, textLookup, func(node *gonx.Node) {
			iterateChildren(node, nodes, textLookup, func(groupNode gonx.Node, groupName string) {
				iterateChildren(&groupNode, nodes, textLookup, func(itemNode gonx.Node, name string) {
					subSearch := base + "/" + groupName + "/" + name + "/info"
					var itm Item
					if !findAndExtract(subSearch, nodes, textLookup, &itm) {
						log.Println("Invalid node search:", subSearch)
						return
					}
					if !addItemByName(name, &itm, items) {
						return
					}
				})
			})
		})
		if !ok {
			log.Println("Invalid node search:", base)
		}
	}

	// Consume has both info and spec: /Item/Consume/<SubGroup>/<ItemID>.img/{info,spec}
	consumeBase := "/Item/Consume"
	ok := gonx.FindNode(consumeBase, nodes, textLookup, func(node *gonx.Node) {
		iterateChildren(node, nodes, textLookup, func(groupNode gonx.Node, groupName string) {
			iterateChildren(&groupNode, nodes, textLookup, func(itemNode gonx.Node, name string) {
				var itm Item
				infoPath := consumeBase + "/" + groupName + "/" + name + "/info"
				if !findAndExtract(infoPath, nodes, textLookup, &itm) {
					log.Println("Invalid node search:", infoPath)
					// continue, spec might still exist but info generally holds core data
				}
				specPath := consumeBase + "/" + groupName + "/" + name + "/spec"
				_ = findAndExtract(specPath, nodes, textLookup, &itm)

				if !addItemByName(name, &itm, items) {
					return
				}
			})
		})
	})
	if !ok {
		log.Println("Invalid node search:", consumeBase)
	}

	// Pet items: /Item/Pet/<ItemID>.img/info
	petBase := "/Item/Pet"
	ok = gonx.FindNode(petBase, nodes, textLookup, func(node *gonx.Node) {
		iterateChildren(node, nodes, textLookup, func(itemNode gonx.Node, name string) {
			subSearch := petBase + "/" + name + "/info"
			var itm Item
			if !findAndExtract(subSearch, nodes, textLookup, &itm) {
				log.Println("Invalid node search:", subSearch)
				return
			}
			// Mark as pet and register
			itm.Pet = true
			if !addItemByName(name, &itm, items) {
				return
			}
		})
	})
	if !ok {
		log.Println("Invalid node search:", petBase)
	}

	return items
}

// iterateChildren abstracts iterating direct children of a node and resolves their names
func iterateChildren(node *gonx.Node, nodes []gonx.Node, textLookup []string, fn func(child gonx.Node, name string)) {
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		child := nodes[node.ChildID+i]
		name := textLookup[child.NameID]
		fn(child, name)
	}
}

// findAndExtract finds a node by path and fills item data using getItem
func findAndExtract(path string, nodes []gonx.Node, textLookup []string, item *Item) bool {
	return gonx.FindNode(path, nodes, textLookup, func(node *gonx.Node) {
		item.getItem(node, nodes, textLookup)
	})
}

// addItemByName parses an item ID from a filename-like name, sets InvTabID and inserts to map
func addItemByName(name string, item *Item, out map[int32]Item) bool {
	trimmed := strings.TrimSuffix(name, filepath.Ext(name))
	id64, err := strconv.ParseInt(trimmed, 10, 32)
	if err != nil {
		log.Println("Invalid item id name:", name, "err:", err)
		return false
	}
	itemID := int32(id64)
	item.InvTabID = byte(itemID / 1e6)
	out[itemID] = *item
	return true
}

func (item *Item) getItem(node *gonx.Node, nodes []gonx.Node, textLookup []string) {

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
			item.IncMDD = float64(gonx.DataToInt16(option.Data))
		case "incPDD":
			item.IncPDD = float64(gonx.DataToInt16(option.Data))
		case "incMAD":
			item.IncMAD = float64(gonx.DataToInt16(option.Data))
		case "incPAD":
			item.IncPAD = float64(gonx.DataToInt16(option.Data))
		case "incEVA":
			item.IncEVA = float64(gonx.DataToInt16(option.Data))
		case "incACC":
			item.IncACC = float64(gonx.DataToInt16(option.Data))
		case "incMHP":
			item.IncMHP = float64(gonx.DataToInt16(option.Data))
		case "recoveryHP":
			item.RecoveryHP = float64(gonx.DataToInt16(option.Data))
		case "incMMP":
			item.IncMMP = float64(gonx.DataToInt16(option.Data))
		case "hp":
			item.HP = gonx.DataToInt16(option.Data)
		case "mp":
			item.MP = gonx.DataToInt16(option.Data)
		case "only":
			item.Only = gonx.DataToInt64(option.Data)
		case "attackSpeed":
			item.AttackSpeed = gonx.DataToInt16(option.Data)
		case "attack":
			item.Attack = float64(gonx.DataToInt16(option.Data))
		case "incSpeed":
			item.IncSpeed = float64(gonx.DataToInt16(option.Data))
		case "incJump":
			item.IncJump = float64(gonx.DataToInt16(option.Data))
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
			item.Path = textLookup[gonx.DataToUint32(option.Data)]
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
		case "moveTo":
			item.MoveTo = gonx.DataToInt32(option.Data)
		case "sample":
		case "iconD":
		case "iconRawD":
		case "iconReward":
		default:
			// Consider gating this log behind a verbosity flag to reduce noise in production.
			log.Println("Unsupported NX item option:", optionName, "->", option.Data)
		}

	}
}
