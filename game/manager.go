package game

import (
	"../communication"
	"errors"
	"fmt"
)

type Manager struct {
	// Players storage
	Players map[int]Player
	// Games storage
	GameServers         map[int]*GameServer
	MessageChannel      chan communication.Message
	CommunicationServer *communication.Server
	ServerActions       Actions
	nextPlayerID        int
	nextGameID          int
}

// Initializes
func ManagerInitialize(communicationServer *communication.Server) (*Manager, error) {
	if communicationServer == nil {
		return nil, errors.New("communication server not initialized")
	}

	var messages chan communication.Message = make(chan communication.Message)
	var players map[int]Player = make(map[int]Player)
	var games map[int]*GameServer = make(map[int]*GameServer)

	var manager Manager = Manager{
		Players:             players,
		GameServers:         games,
		MessageChannel:      messages,
		CommunicationServer: communicationServer,
		nextPlayerID:        1,
		nextGameID:          1,
	}

	// Initialize actions
	errActionsInit := InitializeActions(&manager)

	if errActionsInit != nil {
		msg := fmt.Sprintf("unable to initialize server actions: %s", errActionsInit.Error())
		return nil, errors.New(msg)
	}

	// Add communication channel to communication server
	communicationServer.MessageChannel = messages

	return &manager, nil
}

func ManagerStart(communicationServer *communication.Server, manager *Manager) {
	defer communicationServer.WaitGroup.Done()

	for {
		message := <-communicationServer.MessageChannel
		fmt.Printf("Message: %v\n", message)
		_ = ProcessMessage(manager, &message)
	}
}

func ManagerAddGameServer(manager *Manager, server *GameServer) error {
	if manager == nil {
		return errors.New("cannot add game server to manager: manager is NULL")
	}

	if server == nil {
		return errors.New("cannot add game server to manager: game server is NULL")
	}

	manager.GameServers[server.UID] = server

	return nil

}
