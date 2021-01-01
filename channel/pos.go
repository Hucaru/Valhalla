package channel

import (
	"fmt"
	"math"
)

type pos struct {
	x        int16
	y        int16
	foothold int16
}

func newPos(x, y, foothold int16) pos {
	return pos{x: x, y: y, foothold: foothold}
}

func (d pos) String() string {
	return fmt.Sprintf(" %d(x) %d(y)", d.x, d.y)
}

func (d pos) calcDistanceSquare(v pos) int {
	difx := int(d.x - v.x)
	dify := int(d.y - v.y)
	return (difx * difx) + (dify * dify)
}

func (d pos) CalcDistance(v pos) float64 {
	difx := int(d.x - v.x)
	dify := int(d.y - v.y)
	return math.Sqrt(float64((difx * difx) + (dify * dify)))
}
