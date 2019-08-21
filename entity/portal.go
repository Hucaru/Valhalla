package entity

import "github.com/Hucaru/Valhalla/nx"

type portal struct {
	id          byte
	pos         pos
	name        string
	destFieldID int32
	destName    string
	temporary   bool
}

func createPortalFromType() {

}

func createPortalFromData(p nx.Portal) portal {
	return portal{id: p.ID,
		pos:         pos{x: p.X, y: p.Y},
		name:        p.Pn,
		destFieldID: p.Tm,
		destName:    p.Tn,
		temporary:   false}
}

func (p portal) ID() byte           { return p.id }
func (p portal) Pos() pos           { return p.pos }
func (p portal) DestFieldID() int32 { return p.destFieldID }
func (p portal) DestName() string   { return p.destName }
