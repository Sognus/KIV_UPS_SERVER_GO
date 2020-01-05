package game

import (
	"../communication"
	"errors"
	"fmt"
)

type Manager struct {
	Players map[int]Player
	MessageChannel chan communication.Message
	CommunicationServer *communication.Server
	ServerActions Actions
	nextID int
}

// Initializes
func ManagerInitialize(communicationServer *communication.Server) (*Manager, error) {
	if communicationServer == nil {
		return nil, errors.New("communication server not initialized")
	}

	var messages chan communication.Message = make(chan communication.Message)
	var players map[int]Player = make(map[int]Player)

	var manager Manager = Manager{
		Players:        players,
		MessageChannel: messages,
		CommunicationServer: communicationServer,
		nextID: 1,
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

func ManagerStart(communicationServer *communication.Server,manager *Manager) {
	defer communicationServer.WaitGroup.Done()

	for {
		message := <- communicationServer.MessageChannel
		fmt.Printf("Message: %v\n", message)
		_ = ProcessMessage(manager, &message)
	}
}
