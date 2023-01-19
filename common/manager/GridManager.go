package manager

import (
	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/dataController"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/meta-proto/go/mc_metadata"
	"github.com/Hucaru/Valhalla/mnet"
	"golang.org/x/exp/maps"
	proto2 "google.golang.org/protobuf/proto"
	"runtime"
)

type GridInfo struct {
	GridX    int
	GridY    int
	RegionId int64
}

type GridManager struct {
	grids [][][]ConcurrentMap[int64, *mnet.Client]
	plrs  ConcurrentMap[int64, GridInfo]

	pClients        *ConcurrentMap[int64, *mnet.Client]
	gridChangeQueue *dataController.GridLKQueue
	//gridChangeQueue *dataController.GridLKQueue
}

//func (gridMgr *GridManager) Loop(f <-chan func()) {
//	for {
//		switch <-f {
//
//		}
//	}
//}

func (gridMgr *GridManager) Init(_clients *ConcurrentMap[int64, *mnet.Client], fn func(conn *mnet.Client, msg proto2.Message, msgType int)) {
	gridMgr.grids = make([][][]ConcurrentMap[int64, *mnet.Client], 1)
	gridMgr.plrs = New[GridInfo]()

	columns := (constant.LAND_X2 - constant.LAND_X1) / constant.LAND_VIEW_RANGE
	rows := (constant.LAND_Y2 - constant.LAND_Y1) / constant.LAND_VIEW_RANGE

	regions := constant.RegionMax

	r := make([][][]ConcurrentMap[int64, *mnet.Client], regions)

	for _k := 0; _k < regions; _k++ {
		x := make([][]ConcurrentMap[int64, *mnet.Client], columns)

		for i := 0; i < columns; i++ {
			y := make([]ConcurrentMap[int64, *mnet.Client], rows)

			for j := 0; j < rows; j++ {
				d := New[*mnet.Client]()
				y[j] = d
			}
			x[i] = y
		}

		r[_k] = x
	}
	gridMgr.grids = r
	gridMgr.pClients = _clients
	gridMgr.gridChangeQueue = dataController.NewGridLKQueue()
	go gridMgr.Run(fn)
}

func (gridMgr *GridManager) Add(region int64, gridX, gridY int, cl *mnet.Client) {
	plr := (*cl).GetPlayer()

	gridMgr.grids[region][gridX][gridY].Set(plr.UId, cl)
	gridMgr.plrs.Set(plr.UId, GridInfo{gridX, gridY, region})
}

func (gridMgr *GridManager) Remove(uId int64) *mnet.Client {
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

func (gridMgr *GridManager) FillPlayers(RegionId int64, GridX, GridY int) map[int64]*mnet.Client {
	return gridMgr.fillPlayers(RegionId, GridX, GridY)
}

func (gridMgr *GridManager) fillPlayers(RegionId int64, GridX, GridY int) map[int64]*mnet.Client {
	result := map[int64]*mnet.Client{}

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

	//for v := range gridMgr.grids[RegionId][GridX][GridY].IterBuffered() {
	//	result[v.Key] = v.Val
	//}

	gridMgr.grids[RegionId][GridX][GridY].IterCb(func(k int64, v *mnet.Client) {
		result[k] = v
	})

	return result
}

func (gridMgr *GridManager) TestFunction(oldRegionId, NewRegionId int64, oldX, oldY, newX, newY float32, accountID int64, isNew bool) {
	oldGridX, oldGridY := common.FindGrid(oldX, oldY)
	newGridX, newGridY := common.FindGrid(newX, newY)

	info := dataController.NewGridInfo{
		OldRegionId: oldRegionId,
		NewRegionId: NewRegionId,
		OldGridX:    oldGridX,
		OldGridY:    oldGridY,
		NewGridX:    newGridX,
		NewGridY:    newGridY,
		AccountID:   accountID,
		IsNew:       isNew,
	}

	gridMgr.gridChangeQueue.Enqueue(info)
}

func (gridMgr *GridManager) Run(fn func(conn *mnet.Client, msg proto2.Message, msgType int)) {
	for {
		v := gridMgr.gridChangeQueue.Dequeue()
		if v == nil {
			runtime.Gosched()
			continue
		}

		OldRegionId := v.OldRegionId
		//NewRegionId := v.NewRegionId
		OldGridX := v.OldGridX
		OldGridY := v.OldGridY
		NewGridX := v.NewGridX
		NewGridY := v.NewGridY
		AccountId := v.AccountID

		pClient, ok := gridMgr.pClients.Get(v.AccountID)
		if !ok {
			// emmm....
		}

		p := pClient.GetPlayer_P()
		ch := p.GetCharacter()

		_PlayerInfo := mc_metadata.P2C_PlayerInfo{
			UuId:     p.UId,
			Nickname: ch.NickName,
			Top:      ch.Top,
			Bottom:   ch.Bottom,
			Clothes:  ch.Clothes,
			Hair:     ch.Hair,
		}

		newMe := mc_metadata.P2C_ReportGridNew{
			PlayerInfo: &_PlayerInfo,
			SpawnPosX:  ch.PosX,
			SpawnPosY:  ch.PosY,
			SpawnPosZ:  ch.PosZ,
			SpawnRotX:  ch.RotX,
			SpawnRotY:  ch.RotY,
			SpawnRotZ:  ch.RotZ,
		}

		oldMe := mc_metadata.P2C_ReportGridOld{
			PlayerInfo: &_PlayerInfo,
		}

		if v.IsNew {
			spawnList := map[int64]*mnet.Client{}

			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					_newGridX := NewGridX + i
					_newGridY := NewGridY + j

					maps.Copy(spawnList, gridMgr.fillPlayers(OldRegionId, _newGridX, _newGridY))
				}
			}

			delete(spawnList, AccountId)

			for _, v := range spawnList {
				_p := v.GetPlayer()
				_ch := _p.GetCharacter()
				_PlayerInfo := mc_metadata.P2C_PlayerInfo{
					UuId:     _p.UId,
					Nickname: _ch.NickName,
					Top:      _ch.Top,
					Bottom:   _ch.Bottom,
					Clothes:  _ch.Clothes,
					Hair:     _ch.Hair,
				}

				sP := mc_metadata.P2C_ReportGridNew{
					PlayerInfo: &_PlayerInfo,
					SpawnPosX:  _ch.PosX,
					SpawnPosY:  _ch.PosY,
					SpawnPosZ:  _ch.PosZ,
					SpawnRotX:  _ch.RotX,
					SpawnRotY:  _ch.RotY,
					SpawnRotZ:  _ch.RotZ,
				}

				fn(pClient, &sP, constant.P2C_ReportGridNew)
				fn(v, &newMe, constant.P2C_ReportGridNew)
			}

			continue
		}

		if OldGridX == NewGridX && OldGridY == NewGridY {
			continue
		}

		oldGridList := map[int]GridInfo{}
		newGridList := map[int]GridInfo{}

		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				_oldGridX := OldGridX + i
				_oldGridY := OldGridY + j

				oldGridList[_oldGridX*1000+_oldGridY] = GridInfo{GridX: _oldGridX, GridY: _oldGridY}

				_newGridX := NewGridX + i
				_newGridY := NewGridY + j

				newGridList[_newGridX*1000+_newGridY] = GridInfo{GridX: _newGridX, GridY: _newGridY}
			}
		}

		_newGridList := map[int]GridInfo{}
		_oldGridList := map[int]GridInfo{}

		maps.Copy(_newGridList, newGridList)
		maps.Copy(_oldGridList, oldGridList)

		for k, _ := range oldGridList {
			delete(_newGridList, k)
		}

		for k, _ := range newGridList {
			delete(_oldGridList, k)
		}

		spawnList := map[int64]*mnet.Client{}
		removeList := map[int64]*mnet.Client{}

		for _, v := range _newGridList {
			maps.Copy(spawnList, gridMgr.fillPlayers(OldRegionId, v.GridX, v.GridY))
		}

		for _, v := range _oldGridList {
			maps.Copy(removeList, gridMgr.fillPlayers(OldRegionId, v.GridX, v.GridY))
		}

		delete(removeList, AccountId)
		delete(spawnList, AccountId)

		for _, v := range spawnList {
			_p := v.GetPlayer()
			_ch := _p.GetCharacter()
			_PlayerInfo := mc_metadata.P2C_PlayerInfo{
				UuId:     _p.UId,
				Nickname: _ch.NickName,
				Top:      _ch.Top,
				Bottom:   _ch.Bottom,
				Clothes:  _ch.Clothes,
				Hair:     _ch.Hair,
			}

			sP := mc_metadata.P2C_ReportGridNew{
				PlayerInfo: &_PlayerInfo,
				SpawnPosX:  _ch.PosX,
				SpawnPosY:  _ch.PosY,
				SpawnPosZ:  _ch.PosZ,
				SpawnRotX:  _ch.RotX,
				SpawnRotY:  _ch.RotY,
				SpawnRotZ:  _ch.RotZ,
			}

			fn(pClient, &sP, constant.P2C_ReportGridNew)
			fn(v, &newMe, constant.P2C_ReportGridNew)
		}

		for _, v := range removeList {
			_p := v.GetPlayer()
			_ch := _p.GetCharacter()
			_PlayerInfo := mc_metadata.P2C_PlayerInfo{
				UuId:     _p.UId,
				Nickname: _ch.NickName,
				Top:      _ch.Top,
				Bottom:   _ch.Bottom,
				Clothes:  _ch.Clothes,
				Hair:     _ch.Hair,
			}

			rP := mc_metadata.P2C_ReportGridOld{
				PlayerInfo: &_PlayerInfo,
			}

			fn(pClient, &rP, constant.P2C_ReportGridOld)
			fn(v, &oldMe, constant.P2C_ReportGridOld)
		}
	}

	//for {
	//
	//	addList := map[int64]*mnet.Client{}
	//	removeList := map[int64]*mnet.Client{}
	//
	//	v := gridMgr.gridChangeQueue.Dequeue()
	//	if v == nil {
	//		runtime.Gosched()
	//		continue
	//	}
	//
	//	OldRegionId := v.OldRegionId
	//	NewRegionId := v.NewRegionId
	//	OldGridX := v.OldGridX
	//	OldGridY := v.OldGridY
	//	NewGridX := v.NewGridX
	//	NewGridY := v.NewGridY
	//	AccountId := v.AccountID
	//
	//	log.Println(OldGridX, OldGridY, NewGridX, NewGridY)
	//
	//	conn, _ := gridMgr.pClients.Get(AccountId)
	//
	//	if OldRegionId != NewRegionId {
	//
	//	} else if OldGridX != NewGridX || OldGridY != NewGridY {
	//		_plr := gridMgr.Remove(AccountId)
	//		if _plr != nil {
	//			gridMgr.Add(OldRegionId, NewGridX, NewGridY, _plr)
	//
	//			oldList := map[int64]*mnet.Client{}
	//			newList := map[int64]*mnet.Client{}
	//
	//			for i := -1; i <= 1; i++ {
	//				for j := -1; j <= 1; j++ {
	//					_oldGridX := OldGridX + i
	//					_oldGridY := OldGridY + j
	//
	//					_newGridX := NewGridX + i
	//					_newGridY := NewGridY + j
	//
	//					if !v.IsNew {
	//						maps.Copy(oldList, gridMgr.fillPlayers(OldRegionId, _oldGridX, _oldGridY))
	//					}
	//					maps.Copy(newList, gridMgr.fillPlayers(OldRegionId, _newGridX, _newGridY))
	//				}
	//			}
	//
	//			delete(oldList, AccountId)
	//			delete(newList, AccountId)
	//
	//			for k, v := range oldList {
	//				//_, ok := newList[k]
	//				//if ok {
	//				//	continue
	//				//}
	//
	//				removeList[k] = v
	//			}
	//
	//			for k, v := range newList {
	//				//_, ok := oldList[k]
	//				//if ok {
	//				//	continue
	//				//}
	//
	//				addList[k] = v
	//			}
	//		}
	//	} else if v.IsNew {
	//		gridMgr.Add(OldRegionId, NewGridX, NewGridY, conn)
	//		newList := map[int64]*mnet.Client{}
	//
	//		for i := -1; i <= 1; i++ {
	//			for j := -1; j <= 1; j++ {
	//
	//				_newGridX := NewGridX + i
	//				_newGridY := NewGridY + j
	//
	//				maps.Copy(newList, gridMgr.fillPlayers(OldRegionId, _newGridX, _newGridY))
	//			}
	//		}
	//
	//		addList = newList
	//		delete(addList, AccountId)
	//	}
	//
	//	for _, v := range addList {
	//		c := v.GetPlayer_P().GetCharacter()
	//
	//		__PlayerInfo := mc_metadata.P2C_PlayerInfo{
	//			Nickname: c.NickName,
	//			UuId:     AccountId,
	//			Top:      c.Top,
	//			Bottom:   c.Bottom,
	//			Clothes:  c.Clothes,
	//			Hair:     c.Hair,
	//		}
	//
	//		res := mc_metadata.P2C_ReportGridNew{
	//			PlayerInfo: &__PlayerInfo,
	//			SpawnPosX:  c.PosX,
	//			SpawnPosY:  c.PosY,
	//			SpawnPosZ:  c.PosZ,
	//			SpawnRotX:  c.RotX,
	//			SpawnRotY:  c.RotY,
	//			SpawnRotZ:  c.RotZ,
	//		}
	//
	//		p := conn.GetPlayer_P()
	//		ch := p.GetCharacter()
	//
	//		_PlayerInfo := mc_metadata.P2C_PlayerInfo{
	//			UuId:     p.UId,
	//			Nickname: ch.NickName,
	//			Top:      ch.Top,
	//			Bottom:   ch.Bottom,
	//			Clothes:  ch.Clothes,
	//			Hair:     ch.Hair,
	//		}
	//
	//		res2 := mc_metadata.P2C_ReportGridNew{
	//			PlayerInfo: &_PlayerInfo,
	//			SpawnPosX:  ch.PosX,
	//			SpawnPosY:  ch.PosY,
	//			SpawnPosZ:  ch.PosZ,
	//			SpawnRotX:  ch.RotX,
	//			SpawnRotY:  ch.RotY,
	//			SpawnRotZ:  ch.RotZ,
	//		}
	//
	//		fn(conn, &res, constant.P2C_ReportGridNew)
	//		fn(v, &res2, constant.P2C_ReportGridNew)
	//	}
	//
	//	for k, v := range removeList {
	//		p1 := mc_metadata.P2C_PlayerInfo{
	//			UuId: k,
	//		}
	//
	//		res := mc_metadata.P2C_ReportGridOld{
	//			PlayerInfo: &p1,
	//		}
	//
	//		p2 := mc_metadata.P2C_PlayerInfo{
	//			UuId: conn.GetPlayer().UId,
	//		}
	//
	//		res2 := mc_metadata.P2C_ReportGridOld{
	//			PlayerInfo: &p2,
	//		}
	//
	//		//fmt.Println(fmt.Sprintf("conn : %s v : %s res : %s res2 : %s", conn.GetPlayer().UId, v.GetPlayer().UId, res.PlayerInfo.UuId, res2.PlayerInfo.UuId))
	//
	//		fn(conn, &res, constant.P2C_ReportGridOld)
	//		fn(v, &res2, constant.P2C_ReportGridOld)
	//	}
	//}
}

func (gridMgr *GridManager) OnMove(regionId int64, newX, newY float32, uId int64) (map[int64]*mnet.Client, map[int64]*mnet.Client, map[int64]*mnet.Client) {
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

				oldList := map[int64]*mnet.Client{}
				newList := map[int64]*mnet.Client{}

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

				removeList := map[int64]*mnet.Client{}
				addList := map[int64]*mnet.Client{}

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
			newList := map[int64]*mnet.Client{}
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
