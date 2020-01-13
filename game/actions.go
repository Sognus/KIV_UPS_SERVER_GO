package game

import (
	"../communication"
	"errors"
	"fmt"
	"strconv"
)

type Action func(*Manager, *communication.Message) error

type Actions struct {
	global map[int]Action
	game map[int]Action
}

// action IDs
const (
	// Users request to disconnect
	actionDisconnect = 20
	// Users request to register playername
	actionRegister = 1000
	// Users keepAlive message
	actionKeepAlive = 1100
	// Users request to create game
	actionCreateGame = 2000
	// Users request to join game
	actionJoinGame = 2100
	// Users request to reconnect to existing game
	actionReconnectGame = 2200
	// Users request to list free games
	actionListGames = 2300
	// Users request to get current game status
	actionStatusGame = 2400
)

// Initialize available actions
func InitializeActions(manager *Manager) error {
	manager.ServerActions.global = make(map[int]Action)
	manager.ServerActions.game = make(map[int]Action)

	// RegisterAction messages
	manager.ServerActions.global[actionRegister] = RegisterAction
	manager.ServerActions.global[actionCreateGame] = CreateGameAction
	manager.ServerActions.global[actionJoinGame] = JoinGameAction
	manager.ServerActions.global[actionListGames] = GetGamesListAction
	manager.ServerActions.global[actionDisconnect] = DisconnectAction

	return nil
}

func ProcessMessage(manager *Manager, message *communication.Message) error {
	if manager == nil {
		return errors.New("manager cannot be nil")
	}

	if message == nil {
		return errors.New("message cannot be nil")
	}

	// Create Player if not exist
	player, errPlayerExist := GetPlayerByClientID(manager, message.Source)


	if errPlayerExist != nil {
		errCreatePlayer := CreateUnAuthenticatedPlayer(manager, message.Source)
		if errCreatePlayer != nil {
			return errors.New("unable to create unathenticated player")
		}
	}

	player = player


	// Process global action
	_, ok := manager.ServerActions.global[message.Msg]

	if ok {
		_ = manager.ServerActions.global[message.Msg](manager, message)
	} else {
		fmt.Printf("Client #%d: unknown message (type: %d)\n", message.Source, message.Msg)
		msg := fmt.Sprintf("<id:%d;rid:0;type:10;|status:err;msg:uknown message;>", message.Rid)
		client, errFind := communication.GetClientByID(manager.CommunicationServer, message.Source)

		// SendSocket message to client if client exist
		if errFind == nil {
			_ = communication.SendSocket(manager.CommunicationServer, []byte(msg), client.Socket)
		}

		return errors.New("unknown message")
	}

	return nil
}

// Function to permanently terminate Players account and close connection
func DisconnectAction(manager *Manager, message *communication.Message) error {
	if manager == nil {
		return errors.New("disconnect: manager cannot be nil")
	}

	if message == nil {
		return errors.New("disconnect: message cannot be nil")
	}

	playerIDValue, playerIDPresent := message.Content["playerID"]

	if !playerIDPresent {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:Account not terminated - playerID not provided;>", message.Rid, actionDisconnect))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("could not terminate users account: playerID not provided")
	}

	playerID, errParseInt :=  strconv.Atoi(playerIDValue)

	if errParseInt != nil {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:Account not terminated - playerID must be number;>", message.Rid, actionDisconnect))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("could not terminate users account: playerID is not number")
	}

	player, errFindPlayer := GetPlayerByID(manager, playerID)

	if errFindPlayer != nil || player == nil {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:Account not terminated - player does not exist;>", message.Rid, actionDisconnect))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
	}

	// Terminate client
	if player.client != nil {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:ok;msg:Account terminated;>", message.Rid, actionDisconnect))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		_ = communication.RemoveClient(manager.CommunicationServer, player.client.Socket)
	}

	_ = RemovePlayer(manager, player)
	fmt.Printf("Player #%d: Account and connection terminated\n", playerID)

	return nil
}

// Function to register player's account
func RegisterAction(manager *Manager, message *communication.Message) error {
	if manager == nil {
		return errors.New("register manager cannot be nil")
	}

	if message == nil {
		return errors.New("register: message cannot be nil")
	}

	username, msgNamePresent := message.Content["name"]

	if  !msgNamePresent {
		// Request for registration is missing name send response
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:1000;|status:err;msg:request data are missing;>", message.Rid))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("user could not be registered: request data are missing")
	}

	// Detect if name exist
	nameExist := false
	for _, player := range manager.Players {
		if player.userName == message.Content["name"] {
			nameExist = true
			break
		}
	}

	if nameExist == true {
		// Name was given but name is already used
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:1000;|status:err;msg:username is taken;>", message.Rid))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("user could not be registered: name is already used")
	} else {
		// Check if player exist
		player, errPlayerExist := GetPlayerByClientID(manager, message.Source)

		if errPlayerExist != nil {
			data := []byte(fmt.Sprintf("<id:%d;rid:0;type:1000;|status:err;msg:could not create player;>", message.Rid))
			_ = communication.SendID(manager.CommunicationServer, data, message.Source)
			return errors.New("user could not be registered: player is nil")
		}

		// Check if player is registered
		if isAuthenticated(player) {
			data := []byte(fmt.Sprintf("<id:%d;rid:0;type:1000;|status:err;msg:You cannot register twice;>", message.Rid))
			_ = communication.SendID(manager.CommunicationServer, data, message.Source)
			return errors.New("user could not be registered: cannot register twice")
		}

		player.userName = username
		// User successfully registered
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:1000;|status:ok;msg:user registered;playerID:%d;>", message.Rid, player.ID))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return nil
	}
}

// Function to create new game
func CreateGameAction(manager *Manager, message *communication.Message) error {
	if manager == nil {
		return errors.New("createGame: manager cannot be nil")
	}

	if message == nil {
		return errors.New("createGame: message cannot be nil")
	}

	player, playerFindErr := GetPlayerByClientID(manager, message.Source)

	if playerFindErr != nil {
		return errors.New("player does not exist - WTF")
	}

	if !isAuthenticated(player) {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:2000;|status:error;msg:Game not created - User not registered;>", message.Rid))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("createGame: Player not registered")
	}

	// Check if Player is already in game
	_, errGameExist := GetPlayersGame(manager, player)

	if errGameExist == nil {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:2000;|status:error;msg:Already in another game;>", message.Rid))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("createGame: already in another game")
	}

	// Player exist so we can create game server for him
	gameCreated,  errCreateGame := CreateGame(manager, player)

	if errCreateGame != nil {
		msg := fmt.Sprintf("createGame: %s", errCreateGame.Error())
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:2000;|status:error;msg:Game not created - Create error - %s;>", message.Rid, errCreateGame.Error()))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New(msg)
	}

	// Game was created
	gameCreated.Player1 = player
	data := []byte(fmt.Sprintf("<id:%d;rid:0;type:2000;|status:ok;msg:Game created and joined;GameID:%d;>", message.Rid, gameCreated.UID))
	_ = communication.SendID(manager.CommunicationServer, data, message.Source)

	return nil
}

// Function to join game
func JoinGameAction(manager *Manager, message *communication.Message) error {
	if manager == nil {
		return errors.New("createGame: manager cannot be nil")
	}

	if message == nil {
		return errors.New("createGame: message cannot be nil")
	}

	player, errFindPlayer := GetPlayerByClientID(manager, message.Source)

	if errFindPlayer != nil {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:Game not joined - User does not exist;>", message.Rid, actionJoinGame))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("joinGame: User does not exist")
	}

	// Check if player is registered
	if !isAuthenticated(player) {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:Game not joined - User is not registered;>", message.Rid, actionJoinGame))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("joinGame: User not registered")
	}

	// Check if player is in any game
	_, errConnectedGames := GetPlayersGame(manager, player)

	// No error means player is connected to game
	if errConnectedGames == nil {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:Game not joined - Already have game;>", message.Rid, actionJoinGame))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("joinGame: Player has already game")
	}

	// Check if message content contains game ID key
	gameIDString, idPresent := message.Content["gameID"]

	if !idPresent {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:Game not joined - Missing game ID;>", message.Rid, actionJoinGame))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("joinGame: Missing game ID")
	}

	// Check if game ID is number
	gameID, errParse := strconv.Atoi(gameIDString)

	if errParse != nil {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:Game not joined - Game ID needs to be number;>", message.Rid, actionJoinGame))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("joinGame: Game ID is not number")
	}

	// Check if game exist
	game, errFindGame := GetGameByID(manager, gameID)

	if errFindGame != nil {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:Game not joined - Game with that ID does not exist;>", message.Rid, actionJoinGame))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("joinGame: Game ID is not number")
	}

	// Assign player to game as Player1
	if game.Player1 == nil {
		game.Player1 = player
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:ok;msg:Game #%d joined as Player #1;player:1;>", message.Rid, actionJoinGame, game.UID))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return nil
	}

	// Assign player to game as Player2
	if game.Player2 == nil {
		game.Player2 = player
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:ok;msg:Game #%d joined as Player #2;player:2;>", message.Rid, actionJoinGame, game.UID))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return nil
	}

	data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:Game #%d is FULL;>", message.Rid, actionJoinGame, game.UID))
	_ = communication.SendID(manager.CommunicationServer, data, message.Source)
	return errors.New("joinGame: Game is full")
}

// Function to get GameList
func GetGamesListAction(manager *Manager, message *communication.Message) error {
	if manager == nil {
		return errors.New("createGame: manager cannot be nil")
	}

	if message == nil {
		return errors.New("createGame: message cannot be nil")
	}

	player, errFindPlayer := GetPlayerByClientID(manager, message.Source)

	if errFindPlayer != nil {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:ListGames - User does not exist;>", message.Rid, actionListGames))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("listGames: User does not exist")
	}

	// Check if player is registered
	if !isAuthenticated(player) {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:%d;|status:error;msg:Cannot list games - User is not registered;>", message.Rid, actionListGames))
		_ = communication.SendID(manager.CommunicationServer, data, message.Source)
		return errors.New("listGames: User not registered")
	}

	// Create message base and format
	gameCountFormat := "gameCount:%d;"
	gameIDFormat := "gameID%d:%d;"

	messageBase := fmt.Sprintf("<id:%d;rid:0;type:%d;|", message.Rid, actionListGames)

	// Determine count of empty games
	var empty int = 0
	for _, game := range manager.GameServers {
		if game.Player2 == nil || game.Player1 == nil {
			empty++
		}
	}

	// Build message
	sendMessage := ""
	sendMessage = sendMessage + messageBase
	sendMessage = sendMessage + fmt.Sprintf(gameCountFormat, empty)

	// Build game ids list
	var id int = 0
	for _, game := range manager.GameServers {
		if game.Player2 == nil || game.Player1 == nil {
			sendMessage = sendMessage + fmt.Sprintf(gameIDFormat, id, game.UID)
			id++
		}
	}
	// Add end of message
	sendMessage = sendMessage + ">"

	// Send message to client
	_ = communication.SendID(manager.CommunicationServer, []byte(sendMessage), message.Source)
	return nil
}






