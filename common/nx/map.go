package nx

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Life struct {
	ID      uint32
	Cy      int64
	F       int64
	Fh      int16
	Hide    bool
	MobTime int64
	Rx0     int16
	Rx1     int16
	Npc     bool
	X       int16
	Y       int16
}

type Stage struct {
	Life         []Life
	ForcedReturn uint32
	ReturnMap    uint32
	MobRate      float64
	Town         bool
}

var Maps map[uint32]Stage

func getMapInfo() {
	Maps = make(map[uint32]Stage)
	// Test: Get all map ids
	var maps []string
	// Get the list of maps
	for _, mapSet := range []string{"0", "1", "2", "9"} {
		path := "Map/Map/Map"
		result := searchNode(path+mapSet, func(cursor *node) {
			list := make([]string, int(cursor.ChildCount))

			for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
				n := nodes[cursor.ChildID+i]
				list[i] = path + mapSet + "/" + strLookup[n.NameID]
			}

			maps = append(maps, list...)
		})

		if !result {
			fmt.Println("Bad search:", "Map/Map/Map"+mapSet)
		}
	}
	// Populate the Maps object - Refactor to have more sensible names!
	for _, mapPath := range maps {
		result := searchNode(mapPath, func(cursor *node) {

			dataToInt64 := func(data [8]byte) int64 {
				return int64(data[0]) |
					int64(data[1])<<8 |
					int64(data[2])<<16 |
					int64(data[3])<<24 |
					int64(data[4])<<32 |
					int64(data[5])<<40 |
					int64(data[6])<<48 |
					int64(data[7])<<56
			}

			dataToUint64 := func(data [8]byte) uint64 {
				return uint64(data[0]) |
					uint64(data[1])<<8 |
					uint64(data[2])<<16 |
					uint64(data[3])<<24 |
					uint64(data[4])<<32 |
					uint64(data[5])<<40 |
					uint64(data[6])<<48 |
					uint64(data[7])<<56
			}

			dataToUint32 := func(data [8]byte) uint32 {
				return uint32(data[0]) |
					uint32(data[1])<<8 |
					uint32(data[2])<<16 |
					uint32(data[3])<<24
			}

			dataToInt16 := func(data [8]byte) int16 {
				return int16(data[0]) |
					int16(data[1])<<8
			}

			mapStr := strings.Split(mapPath, "/")
			val, err := strconv.Atoi(strings.Split(mapStr[len(mapStr)-1], ".")[0])

			if err != nil {
				panic(err)
			}

			mapID := uint32(val)
			var lifes node
			var info node

			for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
				mapChild := nodes[cursor.ChildID+i]
				switch strLookup[mapChild.NameID] {
				case "life":
					lifes = mapChild
				case "info":
					info = mapChild
				}
			}
			mapItem := Stage{Life: make([]Life, lifes.ChildCount)}

			// Info handling
			for i := uint32(0); i < uint32(info.ChildCount); i++ {
				n := nodes[info.ChildID+i]

				for j := uint32(0); j < uint32(n.ChildCount); j++ {
					infoNode := nodes[n.ChildID+j]
					switch strLookup[infoNode.NameID] {
					case "forcedReturn":
						mapItem.ForcedReturn = uint32(dataToInt64(infoNode.Data))
					case "mobRate":
						mapItem.MobRate = math.Float64frombits(dataToUint64(infoNode.Data))
					case "returnMap":
						mapItem.ReturnMap = uint32(dataToInt64(infoNode.Data))
					case "town":
						mapItem.Town = bool(dataToInt64(infoNode.Data) == 1)
					}
				}
			}

			// Life handling
			lifeItem := Life{}

			for i := uint32(0); i < uint32(lifes.ChildCount); i++ {
				n := nodes[lifes.ChildID+i]

				for j := uint32(0); j < uint32(n.ChildCount); j++ {
					lifeNode := nodes[n.ChildID+j]

					switch strLookup[lifeNode.NameID] {
					case "id":
						val, err := strconv.Atoi(strLookup[dataToUint32(lifeNode.Data)])

						if err != nil {
							panic(err)
						}

						lifeItem.ID = uint32(val)
					case "cy":
						lifeItem.Cy = dataToInt64(lifeNode.Data)
					case "f":
						lifeItem.F = dataToInt64(lifeNode.Data)
					case "fh":
						lifeItem.Fh = dataToInt16(lifeNode.Data)
					case "hide":
						lifeItem.Hide = bool(dataToInt64(lifeNode.Data) == 1)
					case "mobTime":
						lifeItem.MobTime = dataToInt64(lifeNode.Data)
					case "rx0":
						lifeItem.Rx0 = dataToInt16(lifeNode.Data)
					case "rx1":
						lifeItem.Rx1 = dataToInt16(lifeNode.Data)
					case "type":
						lifeItem.Npc = bool(strLookup[dataToUint32(lifeNode.Data)] == "n")
					case "x":
						lifeItem.X = dataToInt16(lifeNode.Data)
					case "y":
						lifeItem.Y = dataToInt16(lifeNode.Data)
					case "info":
						//
					default:
						fmt.Println("Unkown life type from nx file:", strLookup[lifeNode.NameID], "->", lifeNode.Data)
					}
				}
				mapItem.Life[i] = lifeItem
			}
			Maps[mapID] = mapItem
		})
		if !result {
			fmt.Println("Bad search:", mapPath)
		}
	}
}
