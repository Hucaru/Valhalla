package entity

import (
	"fmt"
	"math"
)

type pos struct {
	x int16
	y int16
}

func (p pos) String() string {
	return fmt.Sprintf(" %d(x) %d(y)", p.x, p.y)
}

func (p pos) calcDistanceSquare(v pos) int {
	difx := int(p.x - v.x)
	dify := int(p.y - v.y)
	return (difx * difx) + (dify * dify)
}

func (p pos) calcDistance(v pos) float64 {
	difx := int(p.x - v.x)
	dify := int(p.y - v.y)
	return math.Sqrt(float64((difx * difx) + (dify * dify)))
}
