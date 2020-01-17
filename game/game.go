package game

import (
	"../communication"
	"errors"
	"time"
)

type GameServer struct {
	// GameServer (Lobby) ID
	UID int
	// Players
	Player1 *Player
	Player2 *Player
	// Ball
	// TODO: add ball
	// Ticks per second
	Tps int
	// Flag if game is running
	Running bool
	// Flag if game is paused
	Paused bool
	// Information when game loop started
	Start time.Time
	// Max tick duration
	TickDuration int64

	// ##################################
	// GAME SPECIFIC VARIABLE

	WIDTH int				// WIDTH OF GAME: 375
	HEIGHT int				// HEIGHT OF GAME: 600
	PLAYER_SIZE_WIDTH int   // SIZE ON X AXIS: 80
	PLAYER_SIZE_HEIGHT int 	// SIZE ON Y AXIS: 3
	PLAYER_SPEED int		// PLAYER SPEED PER TICK: 8
	PLAYER_GAP int			// PLAYER GAP FROM BORDERS: 10

	// ##################################
	// Server messages
	Messages []communication.Message
}

func GameStart(manager *Manager, game *GameServer) {


	game.Start = time.Now()
	game.TickDuration = int64(1000 / game.Tps)

	nextGameTickTime := time.Since(game.Start).Milliseconds()

	for game.Running {
		// If enough time passed from last tick we can do next tick
		for time.Since(game.Start).Milliseconds() >  nextGameTickTime {
			// Processs messages from players, update their position, pause status
			// TODO: implement

			// Update coordinations of ball
			// TODO: implemet

			// Send current state of game to both players
			// TODO: implement

			// Determine next game tick time
			nextGameTickTime += game.TickDuration
		}

	}
}

// Stops and remove empty game
func RemoveEmptyGame(manager *Manager, game *GameServer) error {
	if manager == nil {
		return errors.New("unable to remove game - manager cannot be nil")
	}

	if game == nil {
		return errors.New("unable to remove game - game cannot be nill")
	}

	// Check if game is empty
	if game.Player1 != nil || game.Player2 != nil {
		return errors.New("unable to remove game - game is not empty")
	}

	// Stop game
	game.Running = false

	// Remove game from games list
	delete(manager.GameServers, game.UID)

	// Return OK
	return nil
}

// Creates new game server and stores it in server manager
func CreateGame(manager *Manager, creator *Player) (*GameServer,error) {
	if manager == nil {
		return nil, errors.New("server manager cannot be nil")
	}
	
	if creator == nil {
		return nil, errors.New("could not ")
	}

	/*
		WIDTH int				// WIDTH OF GAME: 375
		HEIGHT int				// HEIGHT OF GAME: 600
		PLAYER_SIZE_WIDTH int   // SIZE ON X AXIS: 80
		PLAYER_SIZE_HEIGHT int 	// SIZE ON Y AXIS: 3
		PLAYER_SPEED int		// PLAYER SPEED PER TICK: 8
		PLAYER_GAP int			// PLAYER GAP FROM BORDERS: 10
	 */
	
	newGame := GameServer{
		UID:     manager.nextGameID,
		Player1: creator,
		Player2: nil,
		Tps:     30,
		Running: true,
		Paused: true,
		// Constants
		WIDTH: 375,
		HEIGHT: 600,
		PLAYER_SIZE_WIDTH: 80,
		PLAYER_SIZE_HEIGHT: 3,
		PLAYER_SPEED: 8,
		PLAYER_GAP: 8,
		// Game messages
		Messages: make([]communication.Message, 0, 10),
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