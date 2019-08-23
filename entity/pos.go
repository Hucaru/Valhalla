package entity

import "fmt"

type pos struct {
	x int16
	y int16
}

func (p pos) String() string {
	return fmt.Sprintf(" %d(x) %d(y)", p.x, p.y)
}
