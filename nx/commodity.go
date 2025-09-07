package nx

import (
	"log"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hucaru/gonx"
)

// StockState values mirrored for client
const (
	StockStateDefault      = 0 // available
	StockStateNotAvailable = 1
)

// Commodity mirrors /Etc/Commodity.img/<index>/
type Commodity struct {
	Index int32 // numeric key under Commodity.img

	// NX fields
	SN       int32
	ItemID   int32
	Count    int32
	Gender   int32
	Period   int32
	OnSale   int32
	Price    int32
	Priority int32
	Class    int32 // optional

	// Derived
	Category   int32 // floor(SN/10_000_000)+1
	StockState int32 // computed at load time
}

type FeaturedKey struct {
	Category byte
	Gender   byte
	Idx      byte
}

// GetCommodities returns the global commodity map keyed by SN.
func GetCommodities() map[int32]Commodity {
	return commodities
}

// GetCommodity returns a single commodity by SN.
func GetCommodity(sn int32) (Commodity, bool) {
	v, ok := commodities[sn]
	return v, ok
}

func GetCommoditySNByItemID(itemID int32) (int32, bool) {
	sn, ok := itemIDToSN[itemID]
	return sn, ok
}

func GetBestSN(category, gender, idx int) int32 {
	var g byte
	switch gender {
	case 0, 1:
		g = byte(gender)
	case 2:
		g = 1
	default:
		g = 0
	}

	if sn, ok := bestItems[FeaturedKey{Category: byte(category), Gender: g, Idx: byte(idx)}]; ok {
		return sn
	}
	return 0
}

func loadBestItems() {
	bestItems = make(map[FeaturedKey]int32)

	type sel struct {
		sn       int32
		onSale   bool
		priority int32
	}

	// Collect candidates per category
	candidates := make(map[byte][]sel, 9)
	for _, c := range GetCommodities() {
		if c.SN == 0 || c.StockState != StockStateDefault {
			continue
		}
		cat := byte(c.Category)
		if cat < 1 || cat > 9 || cat == 9 { // exclude Quest
			continue
		}
		candidates[cat] = append(candidates[cat], sel{
			sn:       c.SN,
			onSale:   c.OnSale != 0,
			priority: c.Priority,
		})
	}

	// Helper: comparison function for ranking
	less := func(a, b sel) bool {
		if a.onSale != b.onSale {
			return a.onSale // on-sale first
		}
		if a.priority != b.priority {
			return a.priority < b.priority // lower priority first
		}
		return a.sn < b.sn // tie-breaker
	}

	// For each category, select top 5 using an insertion-sort into a small buffer
	for cat := byte(1); cat <= 8; cat++ { // 1..8 (exclude 9)
		src := candidates[cat]
		if len(src) == 0 {
			continue
		}

		top := make([]sel, 0, 5)
		for _, s := range src {
			// Find insertion point
			pos := len(top)
			for i := 0; i < len(top); i++ {
				if less(s, top[i]) {
					pos = i
					break
				}
			}
			// Insert at pos
			if pos == len(top) {
				if len(top) < 5 {
					top = append(top, s)
				}
			} else {
				if len(top) < 5 {
					top = append(top, sel{}) // grow
				}
				copy(top[pos+1:], top[pos:])
				top[pos] = s
			}
			// If we exceeded 5, drop the worst (last)
			if len(top) > 5 {
				top = top[:5]
			}
		}

		// Publish mirrored for both genders (internal 0/1)
		for idx := 0; idx < len(top); idx++ {
			bestItems[FeaturedKey{Category: cat, Gender: 0, Idx: byte(idx)}] = top[idx].sn
			bestItems[FeaturedKey{Category: cat, Gender: 1, Idx: byte(idx)}] = top[idx].sn
		}
	}
}

func computeCategory(sn int32) int32 {
	return int32(math.Floor(float64(sn)/10_000_000.0)) + 1
}

func computeStockState(c Commodity) int {
	// Default availability
	state := StockStateDefault

	// If item is unknown in NX, mark not available
	if _, err := GetItem(c.ItemID); err != nil {
		return StockStateNotAvailable
	}

	return state
}

func preferSN(existing Commodity, hasExisting bool, newC Commodity) bool {
	if !hasExisting {
		return true
	}

	if existing.StockState != StockStateDefault && newC.StockState == StockStateDefault {
		return true
	}

	if existing.StockState == newC.StockState {
		if existing.Priority != newC.Priority {
			return newC.Priority < existing.Priority
		}
		return newC.SN < existing.SN
	}
	return false
}

// extractCommodities builds the commodities map by traversing /Etc/Commodity.img
func extractCommodities(nodes []gonx.Node, text []string) map[int32]Commodity {
	const root = "/Etc/Commodity.img"

	out := make(map[int32]Commodity)
	rev := make(map[int32]int32) // ItemID -> SN (preferred)

	ok := gonx.FindNode(root, nodes, text, func(n *gonx.Node) {
		for i := uint32(0); i < uint32(n.ChildCount); i++ {
			dir := nodes[n.ChildID+i]
			raw := text[dir.NameID]
			name := strings.TrimSuffix(raw, filepath.Ext(raw))
			idx64, err := strconv.ParseInt(name, 10, 32)
			if err != nil {
				continue
			}

			c := Commodity{Index: int32(idx64)}
			for j := uint32(0); j < uint32(dir.ChildCount); j++ {
				f := nodes[dir.ChildID+j]
				switch text[f.NameID] {
				case "SN":
					c.SN = gonx.DataToInt32(f.Data)
				case "ItemId":
					c.ItemID = gonx.DataToInt32(f.Data)
				case "Count":
					c.Count = gonx.DataToInt32(f.Data)
				case "Gender":
					c.Gender = gonx.DataToInt32(f.Data)
				case "Period":
					c.Period = gonx.DataToInt32(f.Data)
				case "OnSale":
					c.OnSale = gonx.DataToInt32(f.Data)
				case "Price":
					c.Price = gonx.DataToInt32(f.Data)
				case "Priority":
					c.Priority = gonx.DataToInt32(f.Data)
				case "Class":
					c.Class = gonx.DataToInt32(f.Data)
				}
			}

			if c.SN == 0 || c.ItemID == 0 {
				continue
			}

			c.Category = computeCategory(c.SN)
			c.StockState = int32(computeStockState(c))

			out[c.SN] = c

			if existingSN, ok := rev[c.ItemID]; ok {
				existing := out[existingSN]
				if preferSN(existing, true, c) {
					rev[c.ItemID] = c.SN
				}
			} else {
				rev[c.ItemID] = c.SN
			}
		}
	})
	if !ok {
		log.Println("Invalid node search:", root)
	}

	// Publish globals
	commodities = out
	itemIDToSN = rev

	return out
}
