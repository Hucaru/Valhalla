package game

import (
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/mnet"
)

type Item struct {
	def.Item
	// drop information
	owner                mnet.MConnChannel
	unlockTime, fadeTime int64
}
