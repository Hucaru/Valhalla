package model

type Account struct {
	AccountID   int64
	RegionID    int64
	UId         string
	CharacterID int64
	Role        int64
	NickName    string
	Time        int64
	Hair        string
	Top         string
	Bottom      string
	Clothes     string
	PosX        float32
	PosY        float32
	PosZ        float32
	RotX        float32
	RotY        float32
	RotZ        float32
}
