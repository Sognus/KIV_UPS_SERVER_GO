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
	clients map[int]*Client
	// Wait for all goroutines
	WaitGroup sync.WaitGroup
}

// Prepares server structure to listen
func Init(ip string, port string) (*server, error) {
	fmt.Printf("Server initialization started..\n")

	// Parse address from func argument
	address, errAddr := parsing.AddressFromString(ip, port)

	if errAddr != nil {
		msg := fmt.Sprintf("Unable to initialize server: %s\n", errAddr.Error())
		return nil, errors.New(msg)
	}

	// Create master socket
	masterSocket, errSocket := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)

	// Check for socket error
	if errSocket != nil {
		msg := "Unable to initialize server: Could not create master socket!\n"
		return nil, errors.New(msg)
	}

	// Set socket options to allow address reuse
	errOpts := syscall.SetsockoptInt(masterSocket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if errOpts != nil {
		msg := "Unable to initialize server: Could not set master socket options\n"
		return nil, errors.New(msg)
	}

	// Bind address to socket
	errBind := syscall.Bind(masterSocket, address)

	if errBind != nil {
		msg := fmt.Sprintf("Unable to initialize server: Could not bind address to server\n")
		return nil, errors.New(msg)
	}

	// Start listener
	errListen := syscall.Listen(masterSocket, 5)

	if errListen != nil {
		msg := fmt.Sprintf("Unable to initialize server: Could not start listener!\n")
		return nil, errors.New(msg)
	}

	// Create server context
	serverContext := server{
		masterSocket: masterSocket,
		clients:      make(map[int]*Client),
		WaitGroup:    sync.WaitGroup{},
	}

	// Inform terminal
	fmt.Printf("Server initialization completed\n")

	// Return server context
	return &serverContext, nil
}

// Register new TCP client with server structure
func AddClient(serverContext *server, newClient *Client) error {
	if serverContext == nil {
		return errors.New("Could not register TCP client: server structure is NULL\n")
	}

	// Add Client to storage
	(*serverContext).clients[newClient.Socket] = newClient

	// Start Decoder for client
	go Decoder(newClient)

	return nil
}

// Removes client from server
func RemoveClient(serverContext *server, socketDescriptor int ) error {
	// Check for socket error
	if serverContext == nil {
		return errors.New("Could not remove TCP client: server structure is NULL\n")
	}

	deleteClient, exist := (*serverContext).clients[socketDescriptor]

	// Remove client from storage if it can be removed
	if exist {
		_ = deleteClient.Reader.Close()
		_ = deleteClient.writer.Close()
		_ = syscall.Close(deleteClient.Socket)

		fmt.Printf("Client disconnected: #%d (%s:%d)\n", deleteClient.Socket, deleteClient.ip, deleteClient.port)

		delete((*serverContext).clients, socketDescriptor)
	}

	return nil
}

// Broadcasts message to all connected clients - including sender
func Broadcast(serverContext *server, data []byte) error {
	if serverContext == nil {
		return errors.New("could not broadcast message: server structure is NULL\n")
	}

	for _, client := range (*serverContext).clients {
		_, _ = syscall.Write(client.Socket, data)
	}

	return nil
}

// Broadcasts message for currently connected clients - does not send message to sender
func BroadcastExceptSender(serverContext *server, data []byte, socketSource int) error {
	if serverContext == nil {
		return errors.New("could not broadcast message: server structure is NULL\n")
	}

	for _, client := range (*serverContext).clients {
		if client.Socket != socketSource {
			_, _ = syscall.Write(client.Socket, data)
		}
	}

	return nil
}
