package rectangle

import "math"

// Data of the rectangle
type Data struct {
	Left, Top, Right, Bottom int64
}

// CreateFromLTRB from:
// x-coordinate of the upper-left corner,
// y-coordinate of the upper-left corner,
// x-coordinate of the lower-right corner,
// y-coordinate of the lower-right corner
func CreateFromLTRB(left, top, right, bottom int64) Data {
	return Data{Left: left, Top: top, Right: right, Bottom: bottom}
}

// Inflate the rectangle keeping the geometric centre and return this new rectangle
func (data Data) Inflate(x, y int64) Data {
	xDelta := x / 2
	yDelta := y / 2

	return Data{Left: data.Left - xDelta,
		Top:    data.Top + yDelta,
		Right:  data.Right + xDelta,
		Bottom: data.Bottom - yDelta,
	}
}

// Empty if the rectangle is not initialised/has zero area
func (data Data) Empty() bool {
	if data.Left == 0 && data.Top == 0 && data.Right == 0 && data.Bottom == 0 {
		return true
	}

	return false
}

// Width of rectangle calculated from left - right
func (data Data) Width() int64 {
	return int64(math.Abs(float64(data.Left) - float64(data.Right)))
}

// Height of rectangle calculated from top - bottom
func (data Data) Height() int64 {
	return int64(math.Abs(float64(data.Top) - float64(data.Bottom)))
}
