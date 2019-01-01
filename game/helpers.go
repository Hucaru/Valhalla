package game

// func AddPlayer(player Player) {
// 	players[player.MConnChannel] = player
// 	player.sendMapItems()
// 	maps[player.char.MapID].addController(player.MConnChannel)
// }

// func GetPlayerFromID(id int32) (Player, error) {
// 	for _, p := range players {
// 		if p.Char().ID == id {
// 			return p, nil
// 		}
// 	}

// 	return Player{}, fmt.Errorf("Player ID %i not found", id)
// }

// func GetPlayerFromName(name string) (Player, error) {
// 	for _, p := range players {
// 		if p.Char().Name == name {
// 			return p, nil
// 		}
// 	}

// 	return Player{}, fmt.Errorf("Player name %s not found", name)
// }

// func GetPlayersFromMapID(id int32) []Player {
// 	playerList := []Player{}

// 	for _, v := range players {
// 		if v.char.MapID == id {
// 			playerList = append(playerList, v)
// 		}
// 	}

// 	return playerList
// }
