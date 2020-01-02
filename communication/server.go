package communication

import (
	"errors"
	"fmt"
	"sync"
	"syscall"
)
import "../parsing"

// Server structure
type server struct {
	// Master socket
	masterSocket int
	// Clients
	clients []*client
	// Wait for all goroutines
	WaitGroup sync.WaitGroup
}

// Prepares server structure to listen
func Init(ip string, port string) (*server, error) {
	fmt.Printf("Server initialization started..\n")

	// Create master socket
	masterSocket, errSocket := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)

	// Check for socket error
	if errSocket != nil {
		msg := "Unable to initialize server: Could not create master socket!\n"
		fmt.Printf(msg)
		return nil, errors.New(msg)
	}

	// Parse address from func argument
	address, errAddr := parsing.AddressFromString(ip, port)

	if errAddr != nil {
		msg := fmt.Sprintf("Unable to initialize server: %s\n", errAddr.Error())
		fmt.Printf(msg)
		return nil, errors.New(msg)
	}

	// Bind address to socket
	errBind := syscall.Bind(masterSocket, address)

	if errBind != nil {
		msg := fmt.Sprintf("Unable to initialize server: Could not bind address to server\n")
		return nil, errors.New(msg)
	}

	// Create server context
	serverContext := server{
		masterSocket: masterSocket,
		clients:      make([]*client, 0, 8),
		WaitGroup:    sync.WaitGroup{},
	}

	// Inform terminal
	fmt.Printf("Server initialization completed\n")

	// Return server context
	return &serverContext, nil
}

// Register new TCP client with server structure
func AddClient(serverContext *server, newClient *client) error {
	if serverContext == nil {
		return errors.New("Could not register TCP client: server structure is NULL\n")
	}

	(*serverContext).clients = append((*serverContext).clients, newClient)
	return nil
}

// Removes client from server
func RemoveClient(serverContext *server, socketDescriptor int ) error {
	// Check for socket error
	if serverContext == nil {
		return errors.New("Could not remove TCP client: server structure is NULL\n")
	}

	clients := make([]*client, 0, len((*serverContext).clients))

	for _, client := range (*serverContext).clients {
		if client.socket == socketDescriptor {
			fmt.Printf("Client #%d from %s port %d disconnected\n", client.socket, client.ip, client.port)
			continue
		} else {
			clients = append(clients, client)
		}
	}

	(*serverContext).clients = clients
	return nil
}

// Broadcasts message to all connected clients
func Broadcast(serverContext *server, data []byte) error {
	if serverContext == nil {
		return errors.New("could not broadcast message: server structure is NULL\n")
	}

	for _, client := range (*serverContext).clients {
		_, _ = syscall.Write(client.socket, data)
	}

	return nil
}
