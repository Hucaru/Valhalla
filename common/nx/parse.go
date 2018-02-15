package nx

import (
	"encoding/binary"
	"os"
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

	strLookup := readStrings(f, head)
	nodes := readNodes(f, head)

	constructHierachy(nodes, strLookup)
}

// Currently only interested in the following:
// Character - equips
// Item - all inventory items except equips
// Map - return map is map id to return to upon death, life is mob spawn pos?, town = 1 true, create basic struct for map info
// Mob - simple, find largest obj and create struct for it, get town
// Quest -
// NPC - stand - origin - x,y, take 0 if multiple
func constructHierachy(nodes []node, strLookup []string) {
	// for i, v := range nodes {
	// 	fmt.Println("[", i, "]", strLookup[v.NameID], "with", v.ChildCount, "child nodes ->", v.ChildID)
	//
	// 	if i == 19 {
	// 		break
	// 	}
	//
	// }
}

func readNodes(f *os.File, head header) []node {
	_, err := f.Seek(head.NodeBlockOffset, 0)

	if err != nil {
		panic(err)
	}

	nodes := make([]node, head.NodeCount)
	err = binary.Read(f, binary.LittleEndian, &nodes)

	if err != nil {
		panic(err)
	}

	return nodes
}

func readStrings(f *os.File, head header) []string {
	_, err := f.Seek(head.StringOffsetTableOffset, 0)

	if err != nil {
		panic(err)
	}

	stringOffsets := make([]int64, head.StringCount)
	err = binary.Read(f, binary.LittleEndian, &stringOffsets)

	if err != nil {
		panic(err)
	}

	stringLookUp := make([]string, head.StringCount)

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

		stringLookUp[i] = string(str)
	}

	return stringLookUp
}
