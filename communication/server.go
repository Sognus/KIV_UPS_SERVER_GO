package communication

import (
	"errors"
	"fmt"
	"sync"
	"syscall"
)
import "../parsing"

// Server structure
type Server struct {
	// Master socket
	masterSocket int
	// Clients
	Clients map[int]*Client
	// Destination to send parsed messages to
	MessageChannel chan Message
	// Wait for all goroutines
	WaitGroup sync.WaitGroup
	// Next client ID
	NextClientID int
}

// Prepares Server structure to listen
func Init(ip string, port string) (*Server, error) {
	fmt.Printf("Server initialization started..\n")

	// Parse address from func argument
	address, errAddr := parsing.AddressFromString(ip, port)

	if errAddr != nil {
		msg := fmt.Sprintf("Unable to initialize Server: %s\n", errAddr.Error())
		return nil, errors.New(msg)
	}

	// Create master socket
	masterSocket, errSocket := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)

	// Check for socket error
	if errSocket != nil {
		msg := "Unable to initialize Server: Could not create master socket!\n"
		return nil, errors.New(msg)
	}

	// Set socket options to allow address reuse
	errOpts := syscall.SetsockoptInt(masterSocket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if errOpts != nil {
		msg := "Unable to initialize Server: Could not set master socket options\n"
		return nil, errors.New(msg)
	}

	// Bind address to socket
	errBind := syscall.Bind(masterSocket, address)

	if errBind != nil {
		msg := fmt.Sprintf("Unable to initialize Server: Could not bind address to Server\n")
		return nil, errors.New(msg)
	}

	// Start listener
	errListen := syscall.Listen(masterSocket, 5)

	if errListen != nil {
		msg := fmt.Sprintf("Unable to initialize Server: Could not start listener!\n")
		return nil, errors.New(msg)
	}

	// Create Server context
	serverContext := Server{
		masterSocket:   masterSocket,
		Clients:        make(map[int]*Client),
		WaitGroup:      sync.WaitGroup{},
		MessageChannel: nil,
		NextClientID:   1,
	}

	// Inform terminal
	fmt.Printf("Server initialization completed\n")

	// Return Server context
	return &serverContext, nil
}

// Register new TCP client with Server structure
func AddClient(serverContext *Server, newClient *Client) error {
	if serverContext == nil {
		return errors.New("Could not register TCP client: Server structure is NULL\n")
	}

	// Add Client to storage
	serverContext.Clients[newClient.UID] = newClient

	// Start Decoder for client
	go Decoder(serverContext, newClient)

	return nil
}

// Removes client from Server
func RemoveClient(serverContext *Server, socketDescriptor int) error {
	// Check for socket error
	if serverContext == nil {
		return errors.New("Could not remove TCP client: Server structure is NULL\n")
	}

	deleteClient, errFindClient := GetClientBySocket(serverContext, socketDescriptor)

	// Remove client from storage if it can be removed
	if errFindClient == nil {
		_ = deleteClient.Reader.Close()
		_ = deleteClient.writer.Close()
		_ = syscall.Close(deleteClient.Socket)

		fmt.Printf("Client disconnected: #%d (%s:%d)\n", deleteClient.UID, deleteClient.ip, deleteClient.port)

		delete(serverContext.Clients, deleteClient.UID)
		return nil
	} else {
		return errors.New("client did not exist")
	}

}

// Sends data to client
func SendSocket(serverContext *Server, data []byte, socketSource int) error {
	if serverContext == nil {
		return errors.New("could not broadcast Message: Server structure is NULL\n")
	}

	for _, client := range serverContext.Clients {
		if client.Socket == socketSource {
			_, _ = syscall.Write(client.Socket, data)
			break
		}
	}

	return nil
}

func SendID(serverContext *Server, data []byte, clientID int) error {
	if serverContext == nil {
		return errors.New("could not broadcast Message: Server structure is NULL\n")
	}

	for _, client := range serverContext.Clients {
		if client.UID == clientID {
			_, _ = syscall.Write(client.Socket, data)
			break
		}
	}

	return nil
}

// Broadcasts Message to all connected Clients - including sender
func Broadcast(serverContext *Server, data []byte) error {
	if serverContext == nil {
		return errors.New("could not broadcast Message: Server structure is NULL\n")
	}

	for _, client := range (*serverContext).Clients {
		_, _ = syscall.Write(client.Socket, data)
	}

	return nil
}

// Broadcasts Message for currently connected Clients - does not send Message to sender
func BroadcastExceptSender(serverContext *Server, data []byte, socketSource int) error {
	if serverContext == nil {
		return errors.New("could not broadcast Message: Server structure is NULL\n")
	}

	for _, client := range serverContext.Clients {
		if client.Socket != socketSource {
			_, _ = syscall.Write(client.Socket, data)
		}
	}

	return nil
}

// Returns pointer to client by clients ID (not socket ID)
func GetClientByID(serverContext *Server, seekID int) (*Client, error) {
	if serverContext == nil {
		return nil, errors.New("getClientById: server structure cannot be nill")
	}

	for _, client := range serverContext.Clients {
		if client.UID == seekID {
			return client, nil
		}
	}

	return nil, errors.New("getClientById: client does not exist")
}

func GetClientBySocket(serverContext *Server, socket int) (*Client, error) {
	if serverContext == nil {
		return nil, errors.New("getClientById: server structure cannot be nill")
	}

	for _, client := range serverContext.Clients {
		if client.Socket == socket {
			return client, nil
		}
	}

	return nil, errors.New("getClientById: client does not exist")
}
