package entity

import "github.com/Hucaru/Valhalla/nx"

type Portal struct {
	id          byte
	pos         pos
	name        string
	destFieldID int32
	destName    string
	temporary   bool
}

func createPortalFromType() {

}

func createPortalFromData(p nx.Portal) Portal {
	return Portal{id: p.ID,
		pos:         pos{x: p.X, y: p.Y},
		name:        p.Pn,
		destFieldID: p.Tm,
		destName:    p.Tn,
		temporary:   false}
}

func (p Portal) ID() byte           { return p.id }
func (p Portal) Pos() pos           { return p.pos }
func (p Portal) DestFieldID() int32 { return p.destFieldID }
func (p Portal) DestName() string   { return p.destName }
