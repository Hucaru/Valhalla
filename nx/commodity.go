package nx

import (
	"log"
	"math"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

func GetBestSN(category, gender, idx byte) int32 {
	if sn, ok := bestItems[FeaturedKey{Category: category, Gender: gender, Idx: idx}]; ok {
		return sn
	}
	return 0
}

func loadBestItems() {
	// Build candidate lists per (category, gender).
	type key struct{ cat, gen byte }
	candidates := make(map[key][]int32, 18)

	for _, c := range GetCommodities() {
		if c.SN == 0 {
			continue
		}

		cat := byte(c.Category)
		if cat < 1 || cat > 9 {
			continue
		}

		if c.StockState != StockStateDefault {
			continue
		}

		if c.Gender == 0 {
			k0 := key{cat: cat, gen: 0}
			k1 := key{cat: cat, gen: 1}
			candidates[k0] = append(candidates[k0], c.SN)
			candidates[k1] = append(candidates[k1], c.SN)
			continue
		}

		if c.Gender == 1 || c.Gender == 2 {
			gen := byte(0)
			if c.Gender != 1 {
				gen = 1
			} else {
				gen = 0
			}
			k := key{cat: cat, gen: gen}
			candidates[k] = append(candidates[k], c.SN)
		} else if c.Gender == 0 || c.Gender == 1 {
			g := byte(c.Gender)
			k := key{cat: cat, gen: g}
			candidates[k] = append(candidates[k], c.SN)
		}
	}

	// Randomize selection
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := byte(1); i <= 9; i++ {
		for j := byte(0); j <= 1; j++ {
			list := candidates[key{cat: i, gen: j}]
			if len(list) > 1 {
				rng.Shuffle(len(list), func(a, b int) { list[a], list[b] = list[b], list[a] })
			}
			limit := 5
			if len(list) < limit {
				limit = len(list)
			}

			for k := 0; k < limit; k++ {
				bestItems[FeaturedKey{Category: i, Gender: j, Idx: byte(k)}] = list[k]
			}
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
