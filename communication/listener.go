package communication

import (
	"fmt"
	"io"
	"syscall"
	"unsafe"
)

// Calculate FdSet size
var FdBits = int(unsafe.Sizeof(0) * 8)

// Insert socket to FDSET
func FD_SET(p *syscall.FdSet, socket int) {
	p.Bits[socket/FdBits] |= int64(uint(1) << (uint(socket) % uint(FdBits)))
}

// Checks socket for activity
func FD_ISSET(p *syscall.FdSet, socket int) bool {
	return (p.Bits[socket/FdBits] & int64(uint(1)<<(uint(socket)%uint(FdBits)))) != 0
}

// Clears fdSet
func FD_ZERO(p *syscall.FdSet) {
	for i := range p.Bits {
		p.Bits[i] = 0
	}
}

// Starts server Listener
func Start(serverContext *server) {
	defer (*serverContext).WaitGroup.Done()

	fmt.Printf("Starting server..\n")

	if serverContext == nil {
		fmt.Printf("Could not start server listener: server structure is NULL\n")
		return
	}

	fmt.Printf("Server started!\n")
	fmt.Printf("Accepting connections..\n")

	// Create FdSet
	rfds := &syscall.FdSet{}

	// Endless loop
	for {
		// Clear FdSet
		FD_ZERO(rfds)

		// Insert master set into FdSet
		FD_SET(rfds, (*serverContext).masterSocket)

		// Add connected clients to FdSet
		maxSD := (*serverContext).masterSocket

		for _, client := range (*serverContext).clients {
			FD_SET(rfds, client.Socket)
			if client.Socket > maxSD {
				maxSD = client.Socket
			}
		}

		// Call select syscall to determine active clients
		_, errSelect := syscall.Select(maxSD+1, rfds, nil, nil, nil)

		if errSelect != nil {
			fmt.Printf("Select error: %s\n", errSelect.Error())
			continue
		}

		// Check for master server communication
		if FD_ISSET(rfds, (*serverContext).masterSocket) {
			newSocketDescriptor, newAddress, errAccept := syscall.Accept((*serverContext).masterSocket)

			if errAccept != nil {
				fmt.Printf("Accept error: %s\n", errAccept.Error())
			} else {
				// Get IPv4 address
				address := newAddress.(*syscall.SockaddrInet4)
				ipv4 := fmt.Sprintf("%d.%d.%d.%d", address.Addr[0], address.Addr[1], address.Addr[2], address.Addr[3])
				port := address.Port

				reader, writer := io.Pipe()

				newClient := Client{
					Socket:  newSocketDescriptor,
					ip:      ipv4,
					port:    port,
					address: newAddress,
					Reader:  reader,
					writer:  writer,
				}

				errClientAdd := AddClient(serverContext, &newClient)

				if errClientAdd != nil {
					fmt.Printf(errClientAdd.Error())

				} else {
					// Inform terminal
					fmt.Printf("Client connected: #%d from %s on port %d\n", newClient.Socket, newClient.ip,
						newClient.port)

					// Inform connected client
					msg := "Welcome to Pong server!\n"
					_, _ = syscall.Write(newClient.Socket, []byte(msg))
				}

			}
		}

		// Check for clients socket activity
		for _, client := range (*serverContext).clients {
			// Client activity
			if FD_ISSET(rfds, client.Socket) {
				// Make 64 byte buffer
				buffer := make([]byte, 512)
				// Receive data from socket
				n, errRecv := syscall.Read(client.Socket, buffer)

				if errRecv != nil {
					fmt.Printf("Client #%d: Read error: %s\n", client.Socket, errRecv.Error())
					break
				}

				if n == 0 {
					// Client was disconnected
					_ = RemoveClient(serverContext, client.Socket)
				} else {
					// Write data
					_, _ = client.writer.Write(buffer)
				}

			}
		}
	}
}