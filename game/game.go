package game

import "errors"

type GameServer struct {
	// GameServer (Lobby) ID
	UID int
	// Players
	player1 *Player
	player2 *Player
	// Ticks per second
	Tps int
	// When
	Running bool
}

// Creates new game server and stores it in server manager
func CreateGame(manager *Manager, creator *Player) error {
	if manager == nil {
		return errors.New("server manager cannot be nil")
	}
	
	if creator == nil {
		return errors.New("could not ")
	}
	
	newGame := GameServer{
		UID:     manager.nextGameID,
		player1: creator,
		player2: nil,
		Tps:     30,
		Running: false,
	}

	// Increment next game ID
	manager.nextGameID++

	errAdd := ManagerAddGameServer(manager, &newGame)
	return errAdd
}