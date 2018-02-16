package nx

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type header struct {
	Magic                   [4]byte
	NodeCount               uint32
	NodeBlockOffset         int64
	StringCount             uint32
	StringOffsetTableOffset int64
	BitmapCount             uint32
	BitmapOffsetTableOffset int64
	AudioCount              uint32
	AudioOffsetTableOffset  int64
}

type node struct {
	NameID     uint32
	ChildID    uint32
	ChildCount uint16
	Type       uint16
	Data       [8]byte
}

var strLookup []string
var nodes []node

func Parse(fname string) {
	f, err := os.Open(fname)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	head := header{}
	err = binary.Read(f, binary.LittleEndian, &head)

	if err != nil {
		panic(err)
	}

	if head.Magic != [4]byte{0x50, 0x4B, 0x47, 0x34} {
		panic("Not valid nx magic number")
	}

	readStrings(f, head)
	readNodes(f, head)

	// Test: Get all map ids
	var maps []string

	for _, mapSet := range []string{"0", "1", "2", "9"} {

		result := SearchNode("Map/Map/Map"+mapSet, func(cursor *node) {
			list := make([]string, int(cursor.ChildCount))

			for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
				n := nodes[cursor.ChildID+i]
				list[i] = strings.Split(strLookup[n.NameID], ".")[0]
			}

			maps = append(maps, list...)
		})

		if !result {
			fmt.Println("Bad search:", "Map/Map/Map"+mapSet)
		}
	}
	//fmt.Println(len(maps))
}

// Currently only interested in the following:
// Character - equips
// Item - all inventory items except equips
// Map - return map is map id to return to upon death, life is mob spawn pos?, town = 1 true, create basic struct for map info
// Mob - simple, find largest obj and create struct for it, get town
// Quest -
// NPC - stand - origin - x,y, take 0 if multiple
func SearchNode(search string, fnc func(*node)) bool {
	cursor := &nodes[0]

	path := strings.Split(search, "/")

	if strings.Compare(path[0], "/") == 0 {
		path = path[1:]
	}

	for j, p := range path {
		for i := uint32(0); i < uint32(cursor.ChildCount); i++ {

			if cursor.ChildID+i > uint32(len(nodes))-1 {
				return false
			}

			if strings.Compare(strLookup[nodes[cursor.ChildID+i].NameID], p) == 0 {
				cursor = &nodes[cursor.ChildID+i]

				if j == len(path)-1 {
					fnc(cursor)
					return true
				}

				break
			}
		}
	}

	return false
}

func readNodes(f *os.File, head header) {
	_, err := f.Seek(head.NodeBlockOffset, 0)

	if err != nil {
		panic(err)
	}

	nodes = make([]node, head.NodeCount)
	err = binary.Read(f, binary.LittleEndian, &nodes)

	if err != nil {
		panic(err)
	}
}

func readStrings(f *os.File, head header) {
	_, err := f.Seek(head.StringOffsetTableOffset, 0)

	if err != nil {
		panic(err)
	}

	stringOffsets := make([]int64, head.StringCount)
	err = binary.Read(f, binary.LittleEndian, &stringOffsets)

	if err != nil {
		panic(err)
	}

	strLookup = make([]string, head.StringCount)

	for i, v := range stringOffsets {
		_, err = f.Seek(v, 0)

		if err != nil {
			panic(err)
		}

		var length uint16
		err = binary.Read(f, binary.LittleEndian, &length)

		if err != nil {
			panic(err)
		}

		str := make([]byte, length)
		_, err = f.Read(str)

		if err != nil {
			panic(err)
		}

		strLookup[i] = string(str)
	}
}
