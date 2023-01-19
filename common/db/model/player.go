package model

type Player struct {
	AccountID      int64
	UId            int64
	CharacterID    int64
	RegionID       int64
	character      Character
	interaction    Interaction
	IsBot          int32
	ModifiedAt     int64
	MoveQueueIndex int
}

func (p *Player) GetCharacter() Character {
	return p.character
}

func (p *Player) GetCharacter_P() *Character {
	return &p.character
}

func (p *Player) SetCharacter(character Character) {
	p.character = character
}

func (p *Player) SetInteraction(interaction Interaction) {
	p.interaction = interaction
}

func (p *Player) GetInteraction() Interaction {
	return p.interaction
}

func (p *Player) GetInteraction_P() *Interaction {
	return &p.interaction
}
