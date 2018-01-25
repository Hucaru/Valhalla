package nx

import "os"

type header struct {
	Magic                   uint8
	NodeCount               uint32
	NodeBlockOffset         uint64
	StringCount             uint32
	StringOffsetTableOffset *offsetTable
	BitmapCount             uint32
	BitmapOffsetTableOffset *offsetTable
	AudioCount              uint32
	AudioOffsetTableOffset  *offsetTable
}

type offsetTable struct {
	Offsets []uint64
}

type node struct {
	Name          uint32
	FirstChildID  uint32
	ChildrenCount uint16
	Type          uint16
	DataID        uint32
}

type text struct {
	Length uint16
	Data   uint8
}

type bitmap struct { // We don't need to laod this
	Length uint32
	Data   uint8
}

type audio struct {
	Data uint8
}

type nx struct {
}

var data nx

func Parse(fname string) {
	data = nx{}

	f, err := os.Open(fname)

	if err != nil {
		panic(err)
	}

	defer f.Close()

}
