package pos

import (
	"fmt"
	"math"
)

// Data information that has helpful calculation methods
type Data struct {
	x int16
	y int16
}

// New portal func
func New(x, y int16) Data {
	return Data{x: x, y: y}
}

// X axis value
func (d Data) X() int16 {
	return d.x
}

// SetX axis value
func (d *Data) SetX(v int16) {
	d.x = v
}

// Y axis value
func (d Data) Y() int16 {
	return d.y
}

// SetY axis value
func (d *Data) SetY(v int16) {
	d.y = v
}

// String representation of pos
func (d Data) String() string {
	return fmt.Sprintf(" %d(x) %d(y)", d.x, d.y)
}

// CalcDistanceSquare between two pos (x1 - x2)^2 + (y1 - y2)^2
func (d Data) CalcDistanceSquare(v Data) int {
	difx := int(d.x - v.x)
	dify := int(d.y - v.y)
	return (difx * difx) + (dify * dify)
}

// CalcDistance between two pos sqrt((x1 - x2)^2 + (y1 - y2)^2)
func (d Data) CalcDistance(v Data) float64 {
	difx := int(d.x - v.x)
	dify := int(d.y - v.y)
	return math.Sqrt(float64((difx * difx) + (dify * dify)))
}
