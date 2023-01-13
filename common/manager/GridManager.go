package manager

import (
	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/sasha-s/go-deadlock"
	"golang.org/x/exp/maps"
)

type GridInfo struct {
	GridX int
	GridY int
}

type GridManager struct {
	gridLock deadlock.RWMutex
	grids2   map[int]map[int]map[string]*mnet.Client
	plrs2    map[string]GridInfo
	grids    [][]ConcurrentMap[string, *mnet.Client]
	plrs     ConcurrentMap[string, GridInfo]
}

//func (gridMgr *GridManager) Loop(f <-chan func()) {
//	for {
//		switch <-f {
//
//		}
//	}
//}

func (gridMgr *GridManager) Init() {
	gridMgr.gridLock = deadlock.RWMutex{}
	gridMgr.grids = make([][]ConcurrentMap[string, *mnet.Client], 1)
	gridMgr.plrs = New[GridInfo]()
	gridMgr.grids2 = map[int]map[int]map[string]*mnet.Client{}
	gridMgr.plrs2 = map[string]GridInfo{}

	columns := (constant.LAND_X2 - constant.LAND_X1) / constant.LAND_VIEW_RANGE
	rows := (constant.LAND_Y2 - constant.LAND_Y1) / constant.LAND_VIEW_RANGE

	x := make([][]ConcurrentMap[string, *mnet.Client], columns)

	for i := 0; i < columns; i++ {
		gridMgr.grids2[i] = map[int]map[string]*mnet.Client{}
		y := make([]ConcurrentMap[string, *mnet.Client], rows)

		for j := 0; j < rows; j++ {
			d := New[*mnet.Client]()
			y[j] = d
			gridMgr.grids2[i][j] = map[string]*mnet.Client{}
		}
		x[i] = y
	}

	gridMgr.grids = x
}

func (gridMgr *GridManager) Add(gridX, gridY int, cl *mnet.Client) {
	plr := (*cl).GetPlayer()

	gridMgr.gridLock.Lock()
	gridMgr.grids2[gridX][gridY][plr.UId] = cl
	gridMgr.plrs2[plr.UId] = GridInfo{gridX, gridY}
	gridMgr.gridLock.Unlock()

	//gridMgr.grids[gridX][gridY].Set(plr.UId, cl)
	//gridMgr.plrs.Set(plr.UId, GridInfo{gridX, gridY})
}

func (gridMgr *GridManager) Remove(uId string) *mnet.Client {

	gridMgr.gridLock.RLock()
	info, ok := gridMgr.plrs2[uId]
	if !ok {
		gridMgr.gridLock.RUnlock()
		return nil
	}

	gridMgr.gridLock.RUnlock()

	gridMgr.gridLock.Lock()
	delete(gridMgr.plrs2, uId)

	gridInfo := info
	plr, ok2 := gridMgr.grids2[gridInfo.GridX][gridInfo.GridY][uId]
	if ok2 {
		delete(gridMgr.grids2[gridInfo.GridX][gridInfo.GridY], uId)
		gridMgr.gridLock.Unlock()
		return plr
	}

	gridMgr.gridLock.Unlock()

	//info, ok := gridMgr.plrs.Get(uId)
	//if ok {
	//	gridMgr.plrs.Remove(uId)
	//	gridInfo := info
	//
	//	plr, ok2 := gridMgr.grids[gridInfo.GridX][gridInfo.GridY].Get(uId)
	//	if ok2 {
	//		gridMgr.grids[gridInfo.GridX][gridInfo.GridY].Remove(uId)
	//		return plr
	//	}
	//}

	return nil
}

func (gridMgr *GridManager) FillPlayers(GridX, GridY int) map[string]*mnet.Client {
	return gridMgr.fillPlayers(GridX, GridY)
}

func (gridMgr *GridManager) fillPlayers(GridX, GridY int) map[string]*mnet.Client {
	result := map[string]*mnet.Client{}

	MaxX := (constant.LAND_X2 - constant.LAND_X1) / constant.LAND_VIEW_RANGE
	MaxY := (constant.LAND_Y2 - constant.LAND_Y1) / constant.LAND_VIEW_RANGE

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

	gridMgr.gridLock.RLock()
	result = gridMgr.grids2[GridX][GridY]
	gridMgr.gridLock.RUnlock()

	return result

	//itemChan := gridMgr.grids[GridX][GridY].IterBuffered()
	//for item := range itemChan {
	//	result[item.Key] = item.Val
	//}
	//
	//return result
}

func (gridMgr *GridManager) OnMove(newX, newY float32, uId string) (map[string]*mnet.Client, map[string]*mnet.Client, map[string]*mnet.Client) {
	gridMgr.gridLock.RLock()
	info, ok := gridMgr.plrs2[uId]
	gridMgr.gridLock.RUnlock()
	if ok {
		newGridX, newGridY := common.FindGrid(newX, newY)
		gridInfo := info
		if newGridX != gridInfo.GridX || newGridY != gridInfo.GridY {
			//gridMgr.mtx.Lock()
			//defer gridMgr.mtx.Unlock()

			_plr := gridMgr.Remove(uId)
			if _plr != nil {
				gridMgr.Add(newGridX, newGridY, _plr)

				oldList := map[string]*mnet.Client{}
				newList := map[string]*mnet.Client{}

				for i := -1; i <= 1; i++ {
					for j := -1; j <= 1; j++ {
						oldGridX := gridInfo.GridX + i
						oldGridY := gridInfo.GridY + j

						_newGridX := newGridX + i
						_newGridY := newGridY + j

						maps.Copy(oldList, gridMgr.fillPlayers(oldGridX, oldGridY))
						maps.Copy(newList, gridMgr.fillPlayers(_newGridX, _newGridY))
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

					maps.Copy(newList, gridMgr.fillPlayers(_newGridX, _newGridY))
				}
			}

			delete(newList, uId)

			return nil, nil, newList
		}
	}

	return nil, nil, nil

	//info, ok := gridMgr.plrs.Get(uId)
	//if ok {
	//	newGridX, newGridY := common.FindGrid(newX, newY)
	//	gridInfo := info
	//	if newGridX != gridInfo.GridX || newGridY != gridInfo.GridY {
	//		//gridMgr.mtx.Lock()
	//		//defer gridMgr.mtx.Unlock()
	//
	//		_plr := gridMgr.Remove(uId)
	//		if _plr != nil {
	//			gridMgr.Add(newGridX, newGridY, _plr)
	//
	//			oldList := map[string]*mnet.Client{}
	//			newList := map[string]*mnet.Client{}
	//
	//			for i := -1; i <= 1; i++ {
	//				for j := -1; j <= 1; j++ {
	//					oldGridX := gridInfo.GridX + i
	//					oldGridY := gridInfo.GridY + j
	//
	//					_newGridX := newGridX + i
	//					_newGridY := newGridY + j
	//
	//					maps.Copy(oldList, gridMgr.fillPlayers(oldGridX, oldGridY))
	//					maps.Copy(newList, gridMgr.fillPlayers(_newGridX, _newGridY))
	//				}
	//			}
	//
	//			delete(oldList, uId)
	//			delete(newList, uId)
	//
	//			removeList := map[string]*mnet.Client{}
	//			addList := map[string]*mnet.Client{}
	//
	//			for k, v := range oldList {
	//				_, ok := newList[k]
	//				if ok {
	//					continue
	//				}
	//
	//				removeList[k] = v
	//			}
	//
	//			for k, v := range newList {
	//				_, ok := oldList[k]
	//				if ok {
	//					continue
	//				}
	//
	//				addList[k] = v
	//			}
	//
	//			return addList, removeList, newList
	//		}
	//	} else {
	//		newList := map[string]*mnet.Client{}
	//		for i := -1; i <= 1; i++ {
	//			for j := -1; j <= 1; j++ {
	//				_newGridX := newGridX + i
	//				_newGridY := newGridY + j
	//
	//				maps.Copy(newList, gridMgr.fillPlayers(_newGridX, _newGridY))
	//			}
	//		}
	//
	//		delete(newList, uId)
	//
	//		return nil, nil, newList
	//	}
	//}
	//
	//return nil, nil, nil
}
