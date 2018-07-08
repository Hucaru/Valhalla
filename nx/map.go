package nx

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Life struct {
	ID      int32
	Cy      int64
	F       byte
	Fh      int16
	Hide    bool
	MobTime int64
	Rx0     int16
	Rx1     int16
	IsMob   bool
	X       int16
	Y       int16
}

type Portal struct {
	ID      byte
	Tm      int32
	Tn      string
	Pt      byte
	IsSpawn bool
	X       int16
	Y       int16
	Name    string
}

type Stage struct {
	Life         []Life
	ForcedReturn int32
	ReturnMap    int32
	MobRate      float64
	IsTown       bool
	Portals      []Portal
}

var Maps = make(map[int32]Stage)

func getMapInfo() {
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
			panic("Bad search: Map/Map/Map" + mapSet)
		}
	}
	// Populate the Maps object - Refactor
	for _, mapPath := range maps {
		result := searchNode(mapPath, func(cursor *node) {
			mapStr := strings.Split(mapPath, "/")
			val, err := strconv.Atoi(strings.Split(mapStr[len(mapStr)-1], ".")[0])

			if err != nil {
				panic(err)
			}

			mapID := int32(val)
			var lifes node
			var info node
			var portals node

			for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
				mapChild := nodes[cursor.ChildID+i]
				switch strLookup[mapChild.NameID] {
				case "life":
					lifes = mapChild
				case "info":
					info = mapChild
				case "portal":
					portals = mapChild
				}
			}
			mapItem := Stage{Life: make([]Life, lifes.ChildCount)}

			// Portal handling
			mapItem.Portals = getPortalItem(portals)

			// Info handling
			for i := uint32(0); i < uint32(info.ChildCount); i++ {
				infoNode := nodes[info.ChildID+i]

				switch strLookup[infoNode.NameID] {
				case "forcedReturn":
					mapItem.ForcedReturn = dataToInt32(infoNode.Data)
				case "mobRate":
					mapItem.MobRate = math.Float64frombits(dataToUint64(infoNode.Data))
				case "returnMap":
					mapItem.ReturnMap = dataToInt32(infoNode.Data)
				case "town":
					mapItem.IsTown = bool(infoNode.Data[0] == 1)
				}
			}

			// Life handling
			for i := uint32(0); i < uint32(lifes.ChildCount); i++ {
				mapItem.Life[i] = getLifeItem(nodes[lifes.ChildID+i])
			}

			Maps[mapID] = mapItem
		})
		if !result {
			panic("Bad search:" + mapPath)
		}
	}
}

func getPortalItem(n node) []Portal {
	portals := make([]Portal, n.ChildCount)

	for i := uint32(0); i < uint32(n.ChildCount); i++ {
		p := nodes[n.ChildID+i]
		portal := Portal{}

		portalNumber, err := strconv.Atoi(strLookup[p.NameID])

		if err != nil {
			panic(err)
		}

		portal.ID = byte(portalNumber)

		for j := uint32(0); j < uint32(p.ChildCount); j++ {
			options := nodes[p.ChildID+j]

			switch strLookup[options.NameID] {
			case "pt":
				portal.Pt = options.Data[0]
			case "pn":
				portal.IsSpawn = bool(strLookup[dataToInt32(options.Data)] == "sp")
				portal.Name = strLookup[dataToInt32(options.Data)]
			case "tm":
				portal.Tm = dataToInt32(options.Data)
			case "tn":
				portal.Tn = strLookup[dataToInt32(options.Data)]
			case "x":
				portal.X = dataToInt16(options.Data)
			case "y":
				portal.Y = dataToInt16(options.Data)
			default:
			}
		}

		portals[i] = portal
	}

	return portals
}

func getLifeItem(n node) Life {
	lifeItem := Life{}
	for i := uint32(0); i < uint32(n.ChildCount); i++ {
		lifeNode := nodes[n.ChildID+i]

		switch strLookup[lifeNode.NameID] {
		case "id":
			val, err := strconv.Atoi(strLookup[dataToInt32(lifeNode.Data)])

			if err != nil {
				panic(err)
			}

			lifeItem.ID = int32(val)
		case "cy":
			lifeItem.Cy = dataToInt64(lifeNode.Data)
		case "f":
			lifeItem.F = lifeNode.Data[0]
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
			lifeItem.IsMob = bool(strLookup[dataToInt32(lifeNode.Data)] == "m")
		case "x":
			lifeItem.X = dataToInt16(lifeNode.Data)
		case "y":
			lifeItem.Y = dataToInt16(lifeNode.Data)
		case "info":
			// Don't think this is needed for anythng?
		default:
			fmt.Println("Unkown life type from nx file:", strLookup[lifeNode.NameID], "->", lifeNode.Data)
		}
	}
	return lifeItem
}
