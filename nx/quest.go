package nx

type Quest struct {
	Name   string
	ID     int16
	Items  []int16
	Parent int16
	Order  int16
}
