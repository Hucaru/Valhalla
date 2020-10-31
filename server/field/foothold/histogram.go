package foothold

import (
	"encoding/json"
	"math"

	"github.com/Hucaru/Valhalla/server/pos"
)

// Histogram of foothold data, aims to reduce the amount of footholds that are iterated and compared against one another
// compared to iterating over the slice of all footholds
type Histogram struct {
	footholds []Foothold
	binSize   int
	minX      int16
	bins      [][]*Foothold
}

// CreateHistogram data for a given set of footholds
func CreateHistogram(footholds []Foothold) Histogram {
	minX := footholds[0].x1
	maxX := footholds[0].x2

	for _, v := range footholds[1:] {
		if v.x1 < minX {
			minX = v.x1
		}

		if v.x2 > maxX {
			maxX = v.x2
		}
	}

	delta := maxX - minX
	binSize := int(math.Ceil(float64(delta) / float64(len(footholds))))
	binCount := int(math.Ceil(float64(delta) / float64(binSize)))
	bins := make([][]*Foothold, binCount+1)

	result := Histogram{footholds: footholds, binSize: binSize, minX: minX, bins: bins}

	for i, v := range result.footholds {
		first := result.calculateBinIndex(v.x1)
		last := result.calculateBinIndex(v.x2)

		for j := first; j <= last; j++ {
			result.bins[j] = append(result.bins[j], &result.footholds[i])
		}
	}

	return result
}

func (data Histogram) MarshalJSON() ([]byte, error) {
	bins := make([]int, len(data.bins))

	for i := range bins {
		bins[i] = len(data.bins[i])
	}

	return json.Marshal(struct {
		Bins    []int
		MinX    int16
		BinSize int
	}{
		bins,
		data.minX,
		data.binSize,
	})
}

func (data Histogram) calculateBinIndex(x int16) int {
	ind := x - data.minX

	if ind > 0 {
		ind = int16(math.Ceil(float64(ind) / float64(data.binSize)))
	} else if ind == 0 {
		ind = 0
	} else {
		ind = -1
	}

	return int(ind)
}

// GetFinalPosition from a given position
func (data Histogram) GetFinalPosition(point pos.Data) pos.Data {
	ind := data.calculateBinIndex(point.X())

	if ind < 0 {
		return data.findNearestPoint(0, point)
	} else if ind > len(data.bins)-1 {
		return data.findNearestPoint(len(data.bins)-1, point)
	}

	return data.retrivePosition(ind, point)
}

func (data Histogram) retrivePosition(ind int, point pos.Data) pos.Data {
	minimum := point
	set := false

	for _, v := range data.bins[ind] {
		if !v.Wall() && v.Above(point) {
			pos := v.FindPos(point)

			if pos.Y() >= point.Y() {
				if !set {
					set = true
					minimum = pos
				} else if pos.Y() < minimum.Y() {
					minimum = pos
				}
			}
		}
	}

	return minimum
}

func (data Histogram) findNearestPoint(ind int, point pos.Data) pos.Data {
	return point
}
