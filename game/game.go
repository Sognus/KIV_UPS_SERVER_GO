package game

import (
	"../communication"
	"errors"
	"fmt"
	"time"
)

type GameServer struct {
	// GameServer (Lobby) ID
	UID int
	// Players
	Player1 *Player
	Player2 *Player
	// Ball
	Ball *Ball
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

	Score1 int				// Score for player1
	Score2 int				// Score for player 2

	// ##################################
	// Server messages
	Messages []*communication.Message
	sentMessages int64
}

func GameStart(manager *Manager, game *GameServer) {

	// Initialize game actions


	// Initialize ball
	ball := Ball{
		X:        float64(game.WIDTH / 2),
		Y:        float64(game.HEIGHT / 2),
		Rotation: 45,
		Speed:    5,
		MaxSpeed: 20,
	}
	game.Ball = &ball

	// Setup game tick
	game.Start = time.Now()
	game.TickDuration = int64(1000 / game.Tps)

	nextGameTickTime := time.Since(game.Start).Milliseconds()

	// Start game loop
	for game.Running {
		// If enough time passed from last tick we can do next tick
		for time.Since(game.Start).Milliseconds() >  nextGameTickTime {
			// Process messages from players, update their position, pause status
			// TODO: implement

			// Update coordinations of ball
			_ = UpdateBall(game)

			// Send current state of game to both players
			gameStateMessage, errGameState := BuildGameStateMessage(game)
			// To player 1
			if errGameState == nil && game.Player1 != nil && game.Player1.client != nil {
				_ = communication.SendID(manager.CommunicationServer, []byte(gameStateMessage), game.Player1.client.UID)
			}
			// To player 2
			if errGameState == nil && game.Player2 != nil && game.Player2.client != nil {
				_ = communication.SendID(manager.CommunicationServer, []byte(gameStateMessage), game.Player2.client.UID)
			}



			// Determine next game tick time
			nextGameTickTime += game.TickDuration
		}

	}

	// After end of game send message to players game was completed and delete game
	// TODO: implement
}

// Builds game state message from data
func BuildGameStateMessage(game *GameServer) (string, error) {
	if game == nil {
		return "", errors.New("cannot build game state message: game cannot be null")
	}

	// Start message with message header
	msg := fmt.Sprintf("<id:%d;rid:0;type:2400;|", game.sentMessages)

	// Add player 1 coordiantions
	player1x := int(game.WIDTH / 2)
	player1y := int(0 + game.PLAYER_GAP)
	if game.Player1 != nil {
		player1x = int(game.Player1.x)
		player1y = int(game.Player1.y)
	}

	msg += fmt.Sprintf("player1x:%d;", player1x)
	msg += fmt.Sprintf("player1y:%d;", player1y)

	// Add player 2 coordinations
	player2x := int(game.WIDTH / 2)
	player2y := int(0 + game.PLAYER_GAP)
	if game.Player2 != nil {
		player2x = int(game.Player2.x)
		player2y = int(game.Player2.y)
	}

	msg += fmt.Sprintf("player2x:%d;", player2x)
	msg += fmt.Sprintf("player2y:%d;", player2y)

	// Add score
	msg += fmt.Sprintf("score1:%d;", game.Score1)
	msg += fmt.Sprintf("score2:%d;", game.Score2)

	// Add ball information
	ballX := int(game.WIDTH / 2)
	ballY := int(game.HEIGHT / 2)
	ballSpeed := int(5)
	ballRotation := int(45)
	if game.Ball != nil {
		ballX = int(game.Ball.X)
		ballY = int(game.Ball.Y)
		ballSpeed = game.Ball.Speed
		ballRotation = game.Ball.Rotation
	}

	msg += fmt.Sprintf("ballx:%d;", ballX)
	msg += fmt.Sprintf("bally:%d;", ballY)
	msg += fmt.Sprintf("ballspeed:%d;", ballSpeed)
	msg += fmt.Sprintf("ballrotation:%d;", ballRotation)

	// Add information - is game paused
	msg += fmt.Sprintf("paused:%t;", game.Paused)

	// Add message end
	msg += ">"

	// Increment message sent counter
	game.sentMessages++

	// Return built message
	return msg, nil
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
		// Score
		Score1: 0,
		Score2: 0,
		// Game messages
		Messages: make([]*communication.Message, 0, 10),
		sentMessages: 0,
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