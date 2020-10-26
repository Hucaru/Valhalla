package foothold

import (
	"github.com/Hucaru/Valhalla/server/pos"
)

// Foothold data for a field
type Foothold struct {
	id               int16
	x1, y1, x2, y2   int16
	prev, next       int
	centreX, centreY int16
}

// CreateFoothold from position data
func CreateFoothold(id, x1, y1, x2, y2 int16, prev, next int) Foothold {
	return Foothold{id: id, x1: x1, y1: y1, x2: x2, y2: y2, prev: prev, next: next, centreX: (x2 + x1) / 2, centreY: (y2 + y1) / 2}
}

// Slope if y1 == y2
func (data Foothold) Slope() bool {
	return data.y1 != data.y2
}

// Wall if x1 == x2
func (data Foothold) Wall() bool {
	return data.x1 == data.x2
}

func withinX(check, x1, x2 int16) bool {
	if check >= x1 && check <= x2 {
		return true
	}

	return false
}

func crossProduct(x, x1, x2, y, y1, y2 int16) int16 {
	/*
		cp = |a||b|sin(theta)

		whend dealing with vectors it can be calculated as:

		cp.x = a.y * b.z - a.z * b.y
		cp.y = a.z * b.x - a.x * b.z
		cp.z = a.x * b.y - a.y * b.x

		working in 2d therefore z is zero meaning cp.x & cp.y do not need to be calculated

		since x & y component of cp vector are zero cp.z is the vector magnitude

		|cp| / |a|.|b| = sin(theta)

		if theta lies between 0 and pi crossing pi/2 then it is above the line resulting in a positive value
		if theta lies between 0 and pi crossing 3pi/2 then it is below the line resulting in a negative value
	*/
	return (x-x1)*(y2-y1) - (y-y1)*(x2-x1) // 0 is on the line, > 0 is above, < 0 is below
}

// Above the current foothold
func (data Foothold) Above(p pos.Data) bool {
	if !withinX(p.X(), data.x1, data.x2) {
		return false
	}

	return crossProduct(p.X(), data.x1, data.x2, p.Y(), data.y1, data.y2) >= 0
}

// Below the current foothold
func (data Foothold) Below(p pos.Data) bool {
	if !withinX(p.X(), data.x1, data.x2) {
		return false
	}

	return crossProduct(p.X(), data.x1, data.x2, p.Y(), data.y1, data.y2) <= 0
}

// FootholdBelow below the other foothold
func (data Foothold) FootholdBelow(other Foothold) bool {
	return data.centreY < other.centreY // this method is dodgy I think
}

// FindPos on foothold
func (data Foothold) FindPos(p pos.Data) pos.Data {
	if !data.Slope() {
		return pos.New(p.X(), data.y1, data.id)
	}

	/*
		Equation derived for two collinear points as follows:
		P1 + k(P1 - P2) = R
		x1 + k(x1 - x2) = rx
		k = (rx - x1) / (x1 - x2)

		y1 + k(y1 - y2) = ry

		ry = y1 + ((rx - x1) / (x1 - x2)) * (y1 - y2)

		pre-calculating y1 - y2 and x1 - x2 might yield perf increases (extremely minor)
	*/

	newY := data.y1 + int16((float64(p.X()-data.x1)/float64(data.x1-data.x2))*float64(data.y1-data.y2))

	return pos.New(p.X(), newY, data.id)
}
