package common

import (
	"github.com/Hucaru/Valhalla/constant"
)

func FindGrid(mX float32, mY float32) (int, int) {
	if mX > constant.LAND_X2 {
		mX = constant.LAND_X2
	} else if mX < constant.LAND_X1 {
		mX = constant.LAND_X1
	}

	if mY > constant.LAND_Y2 {
		mY = constant.LAND_Y2
	} else if mY < constant.LAND_Y1 {
		mY = constant.LAND_Y1
	}

	x := (mX - constant.LAND_X1) / constant.LAND_VIEW_RANGE
	y := (mY - constant.LAND_Y1) / constant.LAND_VIEW_RANGE

	return int(x), int(y)
}

func FindLocationInGrid(x1, y1, x2, y2 int) bool {

	if x1 == x2 && y1 == y2 {
		return true
	}

	if x1 == (x2-1) && y1 == (y2+1) {
		return true
	}

	if x1 == x2 && y1 == (y2+1) {
		return true
	}

	if x1 == (x2+1) && y1 == (y2+1) {
		return true
	}

	if x1 == (x2+1) && y1 == y2 {
		return true
	}

	if x1 == (x2+1) && y1 == (y2-1) {
		return true
	}

	if x1 == x2 && y1 == (y2-1) {
		return true
	}

	if x1 == (x2-1) && y1 == (y2-1) {
		return true
	}

	if x1 == (x2-1) && y1 == y2 {
		return true
	}

	return false
}
