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

// GetCommodities returns the global commodity map keyed by SN.
func GetCommodities() map[int32]Commodity {
	return commodities
}

// GetCommodity returns a single commodity by SN.
func GetCommodity(sn int32) (Commodity, bool) {
	v, ok := commodities[sn]
	return v, ok
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

	// Example business rules (match common server behaviors)
	if c.Price == 18000 && c.OnSale != 0 {
		return StockStateNotAvailable
	}

	// If not on sale in data, mark as not available
	if c.OnSale == 0 {
		return StockStateNotAvailable
	}

	return state
}

// extractCommodities builds the commodities map by traversing /Etc/Commodity.img
func extractCommodities(nodes []gonx.Node, text []string) map[int32]Commodity {
	const root = "/Etc/Commodity.img"

	out := make(map[int32]Commodity)

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

			// Only include valid entries (must have SN and ItemID)
			if c.SN == 0 || c.ItemID == 0 {
				continue
			}

			c.Category = computeCategory(c.SN)
			c.StockState = int32(computeStockState(c))

			out[c.SN] = c
		}
	})
	if !ok {
		log.Println("Invalid node search:", root)
	}
	return out
}
