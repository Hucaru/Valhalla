package field

import (
	"github.com/Hucaru/Valhalla/channel/pos"
	"github.com/Hucaru/Valhalla/nx"
)

// Portal that can be plaed in a field
type Portal struct {
	id          byte
	pos         pos.Data
	name        string
	destFieldID int32
	destName    string
	temporary   bool
}

func createPortalFromData(p nx.Portal) Portal {
	return Portal{id: p.ID,
		pos:         pos.New(p.X, p.Y, 0),
		name:        p.Pn,
		destFieldID: p.Tm,
		destName:    p.Tn,
		temporary:   false}
}

// ID of portal
func (p Portal) ID() byte { return p.id }

// Pos the portal takes on the map
func (p Portal) Pos() pos.Data { return p.pos }

// DestFieldID the portal takes the player
func (p Portal) DestFieldID() int32 { return p.destFieldID }

// DestName of the portal on the other side
func (p Portal) DestName() string { return p.destName }
