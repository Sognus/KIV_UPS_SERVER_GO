package game

import (
	"../communication"
	"errors"
)

type Player struct {
	client *communication.Client
	ID int
	userName string
	lastCommunication int64
}

func HasGame(manager *Manager, player *Player) (bool, error) {
	if manager == nil {
		return false, errors.New("manager cannot be NULL")
	}

	if player == nil {
		return false, errors.New("player cannot be NULL")
	}

	for _, playerIter := range manager.Players {
		if playerIter.ID == player.ID {
			// Player was found
			return true, nil
		}
	}

	// Player was not found but thats not error
	return false, nil
}

// Returns Player by his ID
func GetPlayerByID(manager *Manager, playerID int) (*Player, error) {
	if manager == nil {
		return nil, errors.New("manager cannot be NULL")
	}

	for _, playerIter := range manager.Players {
		if playerIter.ID == playerID {
			return &playerIter, nil
		}
	}

	return nil, errors.New("player with that ID does not exist")

}

// Returns Player by his TCP Clients ID
func GetPlayerByClientID(manager *Manager, clientID int) (*Player, error) {
	if manager == nil {
		return nil, errors.New("manager cannot be NULL")
	}

	for _, playerIter := range manager.Players {
		if playerIter.client.UID == clientID {
			return &playerIter, nil
		}
	}

	return nil, errors.New("player with such client ID was not found")
}

// Returns Player Game by Player
func GetPlayersGame(manager *Manager, player *Player) (*GameServer, error) {
	if manager == nil {
		return nil, errors.New("manager cannot be NULL")
	}

	if player == nil {
		return nil, errors.New("player cannot be NULL")
	}

	for _, game := range manager.GameServers {
		// Player is connected as Player1
		if game.Player1 != nil && game.Player1.ID == player.ID {
			return game, nil
		}
		// Player is connected as Player2
		if game.Player2 != nil && game.Player2.ID == player.ID {
			return game, nil
		}
	}

	return nil, errors.New("player is not connected to any game")
}

// Returns Player Game by TCP Client ID
func GetPlayersGameByClientID(manager *Manager, clientID int) (*GameServer, error) {
	if manager == nil {
		return nil, errors.New("manager cannot be NULL")
	}

	player, errFindPlayer := GetPlayerByClientID(manager, clientID)

	if errFindPlayer != nil {
		return nil, errFindPlayer
	}

	game, errFindGame :=  GetPlayersGame(manager, player)

	if errFindGame != nil {
		return nil, errFindGame
	}

	return game, nil
}