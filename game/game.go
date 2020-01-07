package game

import "errors"

type GameServer struct {
	// GameServer (Lobby) ID
	UID int
	// Players
	Player1 *Player
	Player2 *Player
	// Ticks per second
	Tps int
	// When
	Running bool
}

// Creates new game server and stores it in server manager
func CreateGame(manager *Manager, creator *Player) (*GameServer,error) {
	if manager == nil {
		return nil, errors.New("server manager cannot be nil")
	}
	
	if creator == nil {
		return nil, errors.New("could not ")
	}
	
	newGame := GameServer{
		UID:     manager.nextGameID,
		Player1: creator,
		Player2: nil,
		Tps:     30,
		Running: false,
	}

	// Increment next game ID
	manager.nextGameID++

	errAdd := ManagerAddGameServer(manager, &newGame)
	return &newGame,errAdd
}

func GetGameByID(manager *Manager, gameID int) (*GameServer, error) {
	if manager == nil {
		return nil, errors.New("manager cannot be NULL")
	}

	for _, game := range manager.GameServers {
		if game.UID == gameID {
			return game, nil
		}
	}

	return nil, errors.New("game server with that ID does not exist")
}