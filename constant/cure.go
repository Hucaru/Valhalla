package constant

// CureFlag represents flags for debuffs that can be cured by items
type CureFlag int16

const (
	CurePoison   CureFlag = 0x1
	CureWeakness CureFlag = 0x2
	CureCurse    CureFlag = 0x4
	CureDarkness CureFlag = 0x8
	CureSeal     CureFlag = 0x10
)
