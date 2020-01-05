package game

import (
	"../communication"
	"errors"
	"fmt"
	"time"
)

type Action func(*Manager, *communication.Message) error

type Actions struct {
	global map[int]Action
	game map[int]Action
}

// action IDs
const (
	actionRegister = 1000
)

// Initialize available actions
func InitializeActions(manager *Manager) error {
	manager.ServerActions.global = make(map[int]Action)
	manager.ServerActions.game = make(map[int]Action)

	// Register messages
	manager.ServerActions.global[actionRegister] = Register

	return nil
}

func ProcessMessage(manager *Manager, message *communication.Message) error {
	if manager == nil {
		return errors.New("manager cannot be nil")
	}

	if message == nil {
		return errors.New("message cannot be nil")
	}


	// Process global action
	_, ok := manager.ServerActions.global[message.Msg]

	if ok {
		_ = manager.ServerActions.global[message.Msg](manager, message)
	}

	return nil
}

// Function to register player's account
func Register(manager *Manager, message *communication.Message) error {
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
		_ = communication.Send(manager.CommunicationServer, data, message.Source)
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

	// Detect if user is already connected
	var clientConnected *communication.Client = nil
	for _, existPlayer := range manager.Players {
		if existPlayer.client.UID == message.Source {
			clientConnected = existPlayer.client
		}
	}

	if clientConnected != nil {
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:1000;|status:err;msg:multiple accounts not allowed;>", message.Rid))
		_ = communication.Send(manager.CommunicationServer, data, message.Source)
		return errors.New("user could not be registered: client can only be connected once")
	}

	if nameExist == true {
		// Name was given but name is already used
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:1000;|status:err;msg:username is taken;>", message.Rid))
		_ = communication.Send(manager.CommunicationServer, data, message.Source)
		return errors.New("user could not be registered: name is already used")
	} else {
		client, clientExist := manager.CommunicationServer.Clients[message.Source]

		if clientExist == false {
			data := []byte(fmt.Sprintf("<id:%d;rid:0;type:1000;|status:err;msg:unknown tcp connection;>", message.Rid))
			_ = communication.Send(manager.CommunicationServer, data, message.Source)
			return errors.New("user could not be registered: unknown tcp connection")
		}
		
		manager.Players[client.Socket] = Player{
			client:            client,
			ID:                manager.nextID,
			userName:          username,
			lastCommunication: time.Now().Unix(),
		}

		// Increase next ID
		manager.nextID++

		// User successfully registered
		data := []byte(fmt.Sprintf("<id:%d;rid:0;type:1000;|status:ok;msg:user registered;playerID:%d;>", message.Rid, manager.nextID - 1))
		_ = communication.Send(manager.CommunicationServer, data, message.Source)
		return nil
	}
}








