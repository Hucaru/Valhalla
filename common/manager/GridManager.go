package manager

import (
	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
	"golang.org/x/exp/maps"
)

type GridInfo struct {
	GridX    int
	GridY    int
	RegionId int64
}

type GridManager struct {
	grids [][][]ConcurrentMap[string, *mnet.Client]
	plrs  ConcurrentMap[string, GridInfo]
}

//func (gridMgr *GridManager) Loop(f <-chan func()) {
//	for {
//		switch <-f {
//
//		}
//	}
//}

func (gridMgr *GridManager) Init() {
	gridMgr.grids = make([][][]ConcurrentMap[string, *mnet.Client], 1)
	gridMgr.plrs = New[GridInfo]()

	columns := (constant.LAND_X2 - constant.LAND_X1) / constant.LAND_VIEW_RANGE
	rows := (constant.LAND_Y2 - constant.LAND_Y1) / constant.LAND_VIEW_RANGE

	regions := constant.RegionMax

	r := make([][][]ConcurrentMap[string, *mnet.Client], regions)

	for _k := 0; _k < regions; _k++ {
		x := make([][]ConcurrentMap[string, *mnet.Client], columns)

		for i := 0; i < columns; i++ {
			y := make([]ConcurrentMap[string, *mnet.Client], rows)

			for j := 0; j < rows; j++ {
				d := New[*mnet.Client]()
				y[j] = d
			}
			x[i] = y
		}

		r[_k] = x
	}
	gridMgr.grids = r
}

func (gridMgr *GridManager) Add(region int64, gridX, gridY int, cl *mnet.Client) {
	plr := (*cl).GetPlayer()

	gridMgr.grids[region][gridX][gridY].Set(plr.UId, cl)
	gridMgr.plrs.Set(plr.UId, GridInfo{gridX, gridY, region})
}

func (gridMgr *GridManager) Remove(uId string) *mnet.Client {
	info, ok := gridMgr.plrs.Get(uId)
	if ok {
		gridMgr.plrs.Remove(uId)
		gridInfo := info

		plr, ok2 := gridMgr.grids[gridInfo.RegionId][gridInfo.GridX][gridInfo.GridY].Get(uId)
		if ok2 {
			gridMgr.grids[gridInfo.RegionId][gridInfo.GridX][gridInfo.GridY].Remove(uId)
			return plr
		}
	}

	return nil
}

func (gridMgr *GridManager) FillPlayers(RegionId int64, GridX, GridY int) map[string]*mnet.Client {
	return gridMgr.fillPlayers(RegionId, GridX, GridY)
}

func (gridMgr *GridManager) fillPlayers(RegionId int64, GridX, GridY int) map[string]*mnet.Client {
	result := map[string]*mnet.Client{}

	MaxX := (constant.LAND_X2 - constant.LAND_X1) / constant.LAND_VIEW_RANGE
	MaxY := (constant.LAND_Y2 - constant.LAND_Y1) / constant.LAND_VIEW_RANGE

	if 0 > RegionId {
		RegionId = 0
	}

	if RegionId >= constant.RegionMax {
		RegionId = constant.RegionMax - 1
	}

	if 0 > GridX {
		GridX = 0
	}

	if 0 > GridY {
		GridY = 0
	}

	if GridX > MaxX-1 {
		GridX = MaxX - 1
	}

	if GridY > MaxY-1 {
		GridY = MaxY - 1
	}

	gridMgr.grids[RegionId][GridX][GridY].IterCb(func(k string, v *mnet.Client) {
		result[k] = v
	})

	return result
}

func (gridMgr *GridManager) OnMove(regionId int64, newX, newY float32, uId string) (map[string]*mnet.Client, map[string]*mnet.Client, map[string]*mnet.Client) {
	info, ok := gridMgr.plrs.Get(uId)
	if ok {
		newGridX, newGridY := common.FindGrid(newX, newY)
		gridInfo := info
		if gridInfo.RegionId != regionId || newGridX != gridInfo.GridX || newGridY != gridInfo.GridY {
			//gridMgr.mtx.Lock()
			//defer gridMgr.mtx.Unlock()

			_plr := gridMgr.Remove(uId)
			if _plr != nil {
				gridMgr.Add(regionId, newGridX, newGridY, _plr)
				gridMgr.plrs.Set(uId, GridInfo{newGridX, newGridY, regionId})

				oldList := map[string]*mnet.Client{}
				newList := map[string]*mnet.Client{}

				for i := -1; i <= 1; i++ {
					for j := -1; j <= 1; j++ {
						oldGridX := gridInfo.GridX + i
						oldGridY := gridInfo.GridY + j

						_newGridX := newGridX + i
						_newGridY := newGridY + j

						maps.Copy(oldList, gridMgr.fillPlayers(gridInfo.RegionId, oldGridX, oldGridY))
						maps.Copy(newList, gridMgr.fillPlayers(regionId, _newGridX, _newGridY))
					}
				}

				delete(oldList, uId)
				delete(newList, uId)

				removeList := map[string]*mnet.Client{}
				addList := map[string]*mnet.Client{}

				for k, v := range oldList {
					_, ok := newList[k]
					if ok {
						continue
					}

					removeList[k] = v
				}

				for k, v := range newList {
					_, ok := oldList[k]
					if ok {
						continue
					}

					addList[k] = v
				}

				return addList, removeList, newList
			}
		} else {
			newList := map[string]*mnet.Client{}
			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					_newGridX := newGridX + i
					_newGridY := newGridY + j

					maps.Copy(newList, gridMgr.fillPlayers(regionId, _newGridX, _newGridY))
				}
			}

			delete(newList, uId)

			return nil, nil, newList
		}
	}

	return nil, nil, nil
}
