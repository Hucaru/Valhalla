package channel

type maplePortal struct {
	toMap    int32
	toPortal string
	name     string
	x, y     int16
	isSpawn  bool
}

func (m *maplePortal) GetToMap() int32               { return m.toMap }
func (m *maplePortal) SetToMap(mapID int32)          { m.toMap = mapID }
func (m *maplePortal) GetToPortal() string           { return m.toPortal }
func (m *maplePortal) SetToPortal(portalName string) { m.toPortal = portalName }
func (m *maplePortal) GetName() string               { return m.name }
func (m *maplePortal) SetName(name string)           { m.name = name }
func (m *maplePortal) GetX() int16                   { return m.x }
func (m *maplePortal) SetX(x int16)                  { m.x = x }
func (m *maplePortal) GetY() int16                   { return m.y }
func (m *maplePortal) SetY(y int16)                  { m.y = y }
func (m *maplePortal) GetIsSpawn() bool              { return m.isSpawn }
func (m *maplePortal) SetIsSpawn(isSpawn bool)       { m.isSpawn = isSpawn }
