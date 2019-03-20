package nx

import (
	"fmt"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hucaru/gonx"
)

// Portal object in a map
type Portal struct {
	ID     byte
	Pn     string
	Tm     int32
	Tn     string
	Pt     int64
	X, Y   int16
	Script string
}

// Life object in a map
type Life struct {
	ID       int32
	Type     string
	Foothold int16
	FaceLeft bool
	X, Y     int16
	MobTime  int64
	Hide     int64
	Rx0, Rx1 int16
	Cy       int64
	Info     int64
}

// Reactor object in a map
type Reactor struct {
	ID          int64
	FaceLeft    int64
	X, Y        int64
	ReactorTime int64
	Name        string
}

// Foothold in map
type Foothold struct {
	x1, x2, y1, y2 int
}

// Rectangle type for MBR
type Rectangle struct {
	left, top, right, bottom int
}

func (r Rectangle) Empty() bool {
	if (r.left | r.top | r.right | r.bottom) == 0 {
		return true
	}

	return false
}

func (r *Rectangle) Inflate(x, y int) {
	r.left -= x
	r.top += y
	r.right += x
	r.bottom -= y
}

func (r Rectangle) Width() int {
	return r.right - r.left
}

func (r Rectangle) Height() int {
	return r.top - r.bottom
}

// Map data from nx
type Map struct {
	Town         int64
	ForcedReturn int64
	ReturnMap    int32
	MobRate      float64

	Swim, PersonalShop, EntrustedShop, ScrollDisable int64

	MoveLimit int64
	DecHP     int64

	NPCs      []Life
	Mobs      []Life
	Portals   []Portal
	Reactors  []Reactor
	Footholds []Foothold

	FieldLimit                       int64
	VRRight, VRTop, VRLeft, VRBottom int64

	Recovery                  float64
	Version                   int64
	Bgm, MapMark              string
	Cloud, HideMinimap        int64
	MapDesc, Effect           string
	Fs                        float64
	TimeLimit                 int64
	FieldType                 int64
	Everlast, Snow, Rain      int64
	MapName, StreetName, Help string

	VRLimit, MBR, OMBR             Rectangle
	MobCapacityMin, MobCapacityMax int
}

func extractMaps(nodes []gonx.Node, textLookup []string) map[int32]Map {
	maps := make(map[int32]Map)

	searches := []string{"/Map/Map/Map0", "/Map/Map/Map1", "/Map/Map/Map2", "/Map/Map/Map9"}

	for _, search := range searches {
		valid := gonx.FindNode(search, nodes, textLookup, func(node *gonx.Node) {
			for i := uint32(0); i < uint32(node.ChildCount); i++ {
				mapNode := nodes[node.ChildID+i]
				name := textLookup[mapNode.NameID]

				var mapItem Map

				valid := gonx.FindNode(search+"/"+name+"/info", nodes, textLookup, func(node *gonx.Node) {
					mapItem = getMapInfo(node, nodes, textLookup)
				})

				if !valid {
					log.Println("Invalid node search:", search)
				}

				gonx.FindNode(search+"/"+name+"/life", nodes, textLookup, func(node *gonx.Node) {
					mapItem.NPCs, mapItem.Mobs = getMapLifes(node, nodes, textLookup)
				})

				gonx.FindNode(search+"/"+name+"/portal", nodes, textLookup, func(node *gonx.Node) {
					mapItem.Portals = getMapPortals(node, nodes, textLookup)
				})

				gonx.FindNode(search+"/"+name+"/reactor", nodes, textLookup, func(node *gonx.Node) {
					mapItem.Reactors = getMapReactors(node, nodes, textLookup)
				})

				gonx.FindNode(search+"/"+name+"/foothold", nodes, textLookup, func(node *gonx.Node) {
					mapItem.Footholds = getMapFootholds(node, nodes, textLookup)
				})

				name = strings.TrimSuffix(name, filepath.Ext(name))
				mapID, err := strconv.Atoi(name)

				if err != nil {
					log.Println(err)
					continue
				}

				mapItem.calculateMapLimits()

				maps[int32(mapID)] = mapItem
			}
		})

		if !valid {
			log.Println("Invalid node search:", search)
		}

	}

	return maps
}

func (m *Map) calculateMapLimits() {
	left, top := math.MaxInt32, math.MaxInt32
	right, bottom := math.MinInt32, math.MinInt32

	for _, fh := range m.Footholds {
		if fh.x1 < left {
			left = fh.x1
		}

		if fh.y1 < top {
			top = fh.y1
		}

		if fh.x2 < left {
			left = fh.x2
		}

		if fh.y2 < top {
			top = fh.y2
		}

		if fh.x1 > right {
			right = fh.x1
		}

		if fh.y1 > bottom {
			bottom = fh.y1
		}

		if fh.x2 > right {
			right = fh.x2
		}

		if fh.y2 > bottom {
			bottom = fh.y2
		}
	}

	m.VRLimit = Rectangle{int(m.VRLeft), int(m.VRTop), int(m.VRRight), int(m.VRBottom)}

	if m.VRLimit.Empty() {
		m.VRLimit.left, m.VRLimit.top, m.VRLimit.right, m.VRLimit.bottom = left, top-300, right, bottom+75
	}

	left += 30
	top -= 300
	right -= 30
	bottom += 10

	if !m.VRLimit.Empty() {
		if m.VRLimit.left+20 < left {
			left = m.VRLimit.left + 20
		}

		if m.VRLimit.top+65 < top {
			top = m.VRLimit.top + 65
		}

		if m.VRLimit.right-5 > right {
			right = m.VRLimit.right - 5
		}

		if m.VRLimit.bottom > bottom {
			bottom = m.VRLimit.bottom
		}
	}

	m.MBR.left, m.MBR.top, m.MBR.right, m.MBR.bottom = left+10, top-375, right-10, bottom+60
	m.MBR.Inflate(10, 10)
	m.OMBR = m.MBR
	m.OMBR.Inflate(60, 60)

	mobX, mobY := 800, 600

	if m.MBR.Width() > 800 {
		mobX = m.MBR.Width()
	}

	if m.MBR.Height() > 800 {
		mobY = m.MBR.Height()
	}

	m.MobCapacityMin = int((float64(mobX*mobY) * m.MobRate) * 0.0000078125)

	if m.MobCapacityMin < 1 {
		m.MobCapacityMin = 1
	} else if m.MobCapacityMin > 40 {
		m.MobCapacityMin = 40
	}

	m.MobCapacityMax = m.MobCapacityMin * 2
}

func getMapInfo(node *gonx.Node, nodes []gonx.Node, textLookup []string) Map {
	var m Map
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		option := nodes[node.ChildID+i]
		optionName := textLookup[option.NameID]

		switch optionName {
		case "town":
			m.Town = gonx.DataToInt64(option.Data)
		case "mobRate":
			m.MobRate = gonx.DataToFloat64(option.Data)
		case "forcedReturn":
			m.ForcedReturn = gonx.DataToInt64(option.Data)
		case "personalShop":
			m.PersonalShop = gonx.DataToInt64(option.Data)
		case "entrustedShop":
			m.EntrustedShop = gonx.DataToInt64(option.Data)
		case "swim":
			m.Swim = gonx.DataToInt64(option.Data)
		case "moveLimit":
			m.MoveLimit = gonx.DataToInt64(option.Data)
		case "decHP":
			m.DecHP = gonx.DataToInt64(option.Data)
		case "scrollDisable":
			m.ScrollDisable = gonx.DataToInt64(option.Data)
		case "fieldLimit": // Max number of mobs on map?
			m.FieldLimit = gonx.DataToInt64(option.Data)
		// Are VR settings to do with mob spawning? Determine which mob to spawn?
		case "VRRight":
			m.VRRight = gonx.DataToInt64(option.Data)
		case "VRTop":
			m.VRTop = gonx.DataToInt64(option.Data)
		case "VRLeft":
			m.VRLeft = gonx.DataToInt64(option.Data)
		case "VRBottom":
			m.VRBottom = gonx.DataToInt64(option.Data)
		// case "VRLimit":
		// 	m.VRLimit = gonx.DataToInt64(option.Data)
		case "recovery": // float64
			m.Recovery = gonx.DataToFloat64(option.Data)
		case "returnMap":
			m.ReturnMap = gonx.DataToInt32(option.Data)
		case "version":
			m.Version = gonx.DataToInt64(option.Data)
		case "bgm":
			m.Bgm = textLookup[gonx.DataToUint32(option.Data)]
		case "mapMark":
			m.MapMark = textLookup[gonx.DataToUint32(option.Data)]
		case "cloud":
			m.Cloud = gonx.DataToInt64(option.Data)
		case "hideMinimap":
			m.HideMinimap = gonx.DataToInt64(option.Data)
		case "mapDesc":
			m.MapDesc = textLookup[gonx.DataToUint32(option.Data)]
		case "effect":
			m.Effect = textLookup[gonx.DataToUint32(option.Data)]
		case "fs":
			m.Fs = gonx.DataToFloat64(option.Data)
		case "timeLimit": // is this for maps where a user can only be in there for x time?
			m.TimeLimit = gonx.DataToInt64(option.Data)
		case "fieldType":
			m.FieldType = gonx.DataToInt64(option.Data)
		case "everlast":
			m.Everlast = gonx.DataToInt64(option.Data)
		case "snow":
			m.Snow = gonx.DataToInt64(option.Data)
		case "rain":
			m.Rain = gonx.DataToInt64(option.Data)
		case "mapName":
			m.MapName = textLookup[gonx.DataToUint32(option.Data)]
		case "streetName":
			m.StreetName = textLookup[gonx.DataToUint32(option.Data)]
		case "help":
			m.Help = textLookup[gonx.DataToUint32(option.Data)]
		default:
			log.Println("Unsupported NX map option:", optionName, "->", option.Data)
		}
	}

	return m
}

func getMapPortals(node *gonx.Node, nodes []gonx.Node, textLookup []string) []Portal {
	portals := make([]Portal, node.ChildCount)

	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		portalObj := nodes[node.ChildID+i]

		portalNumber, err := strconv.Atoi(textLookup[portalObj.NameID])

		if err != nil {
			fmt.Println("Skiping portal as ID is not a number")
			continue
		}

		portal := Portal{ID: byte(portalNumber)}

		for j := uint32(0); j < uint32(portalObj.ChildCount); j++ {
			option := nodes[portalObj.ChildID+j]
			optionName := textLookup[option.NameID]

			switch optionName {
			case "pt":
				portal.Pt = gonx.DataToInt64(option.Data)
			case "pn":
				portal.Pn = textLookup[gonx.DataToUint32(option.Data)]
			case "tm":
				portal.Tm = gonx.DataToInt32(option.Data)
			case "tn":
				portal.Tn = textLookup[gonx.DataToUint32(option.Data)]
			case "x":
				portal.X = gonx.DataToInt16(option.Data)
			case "y":
				portal.Y = gonx.DataToInt16(option.Data)
			case "script":
				portal.Script = textLookup[gonx.DataToUint32(option.Data)]
			default:
				fmt.Println("Unsupported NX portal option:", optionName, "->", option.Data)
			}
		}

		portals[i] = portal
	}

	return portals
}

func getMapLifes(node *gonx.Node, nodes []gonx.Node, textLookup []string) ([]Life, []Life) {
	npcs := []Life{}
	mobs := []Life{}

	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		lifeObj := nodes[node.ChildID+i]

		var life Life

		for j := uint32(0); j < uint32(lifeObj.ChildCount); j++ {
			option := nodes[lifeObj.ChildID+j]
			optionName := textLookup[option.NameID]

			switch optionName {
			case "id":
				id := textLookup[gonx.DataToUint32(option.Data)]

				tmpID, err := strconv.Atoi(id)

				if err != nil {
					continue
				}

				life.ID = int32(tmpID)
			case "type":
				life.Type = textLookup[gonx.DataToUint32(option.Data)]
			case "fh":
				life.Foothold = gonx.DataToInt16(option.Data)
			case "f":
				life.FaceLeft = gonx.DataToBool(option.Data[0])
			case "x":
				life.X = gonx.DataToInt16(option.Data)
			case "y":
				life.Y = gonx.DataToInt16(option.Data)
			case "mobTime":
				life.MobTime = gonx.DataToInt64(option.Data)
			case "hide":
				life.Hide = gonx.DataToInt64(option.Data)
			case "rx0":
				life.Rx0 = gonx.DataToInt16(option.Data)
			case "rx1":
				life.Rx1 = gonx.DataToInt16(option.Data)
			case "cy":
				life.Cy = gonx.DataToInt64(option.Data)
			case "info": // An npc in map 103000002.img has info field
				life.Info = gonx.DataToInt64(option.Data)
			default:
				fmt.Println("Unsupported NX life option:", optionName, "->", option.Data)
			}
		}

		if life.Type == "m" {
			mobs = append(mobs, life)
		} else if life.Type == "n" {
			npcs = append(npcs, life)
		} else {
			fmt.Println("Unsupported life type:", life.Type)
		}
	}

	return npcs, mobs
}

func getMapReactors(node *gonx.Node, nodes []gonx.Node, textLookup []string) []Reactor {
	reactors := make([]Reactor, node.ChildCount)

	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		reactorObj := nodes[node.ChildID+i]

		var reactor Reactor

		for j := uint32(0); j < uint32(reactorObj.ChildCount); j++ {
			option := nodes[reactorObj.ChildID+j]
			optionName := textLookup[option.NameID]

			switch optionName {
			case "id":
				reactor.ID = gonx.DataToInt64(option.Data)
			case "x":
				reactor.X = gonx.DataToInt64(option.Data)
			case "y":
				reactor.Y = gonx.DataToInt64(option.Data)
			case "f":
				reactor.FaceLeft = gonx.DataToInt64(option.Data)
			case "reactorTime":
				reactor.ReactorTime = gonx.DataToInt64(option.Data)
			case "name":
				reactor.Name = textLookup[gonx.DataToUint32(option.Data)] // boss, ludigate
			default:
				fmt.Println("Unsupported NX reactor option:", optionName, "->", option.Data)
			}
		}
	}

	return reactors
}

func getMapFootholds(node *gonx.Node, nodes []gonx.Node, textLookup []string) []Foothold {
	footholds := []Foothold{}
	return footholds
}
