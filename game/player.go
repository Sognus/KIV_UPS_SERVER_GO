package game

import (
	"../communication"
	"errors"
	"fmt"
	"time"
)

type Player struct {
	// Players communication client
	client *communication.Client
	// Players ID
	ID int
	// Players name
	userName string
	// Unix time of players last message
	lastCommunication int64

	/* Game specific variables */
	/* ####################### */

	// x coordination of player center
	x float64
	// y coordianation of player center
	y float64
	// Player height
	height float64
	// Player width
	width float64
}


func CreateUnAuthenticatedPlayer(manager *Manager, clientID int) error {
	if manager == nil {
		return errors.New("unable to create empty player, manager is null")
	}

	client, errFindClient := communication.GetClientByID(manager.CommunicationServer, clientID)

	if errFindClient != nil {
		return errors.New("unable to create empty player, client does not exist")
	}

	player := Player{
		client:            client,
		ID:                manager.nextPlayerID,
		userName:          "",
		lastCommunication: time.Now().Unix(),
	}

	// Increment new player ID
	manager.nextPlayerID++

	fmt.Printf("New player created: ID #%d\n", player.ID)

	return ManagerAddPlayer(manager, &player)


}


// Returns if player is auth
func isAuthenticated(player *Player) bool {
	if player == nil {
		return false
	}

	return player.userName != ""
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
			return playerIter, nil
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
			return playerIter, nil
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

// Remove player without terminating client by players ID
func RemovePlayerByID(manager *Manager, playerID int) error {
	if manager == nil {
		return errors.New("manager cannot be null")
	}

	player, errFindPlayer := GetPlayerByID(manager, playerID)

	if errFindPlayer != nil {
		return errors.New("player does not exist")
	}

	return RemovePlayer(manager, player)
}
// Remove player without removing client
func RemovePlayer(manager *Manager, player *Player) error {
	if manager == nil {
		return errors.New("manager cannot be null")
	}

	if player == nil {
		return errors.New("player cannot be null")
	}

	delete(manager.Players, player.ID)
	return nil
}