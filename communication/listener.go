package communication

import (
	"bufio"
	"fmt"
	"io"
	"syscall"
	"time"
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

	errListen := syscall.Listen((*serverContext).masterSocket, 8)

	if errListen != nil {
		msg := fmt.Sprintf("Could not start listener: %s\n", errListen.Error())
		fmt.Printf(msg)
		return
	}

	fmt.Printf("Server started!\n")

	// Create FdSet
	rfds := &syscall.FdSet{}
	timeout := &syscall.Timeval{}

	// Endless loop
	for {
		// Clear FdSet
		FD_ZERO(rfds)

		// Insert master set into FdSet
		FD_SET(rfds, (*serverContext).masterSocket)

		// Add connected clients to FdSet
		max_sd := -1
		for _, client := range (*serverContext).clients {
			FD_SET(rfds, client.socket)
			if client.socket > max_sd {
				max_sd = client.socket
			}
		}

		// Call select syscall to determine active clients
		_, errSelect := syscall.Select(max_sd+1, rfds, nil, nil, timeout)

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
				bufferReader := bufio.NewReader(reader)
				bufferWriter := bufio.NewWriter(writer)

				newClient := client{
					socket:  newSocketDescriptor,
					ip:      ipv4,
					port:    port,
					address: newAddress,
					buffer: bufio.NewReadWriter(bufferReader, bufferWriter),
					reader: reader,
					writer: writer,
				}

				errClientAdd := AddClient(serverContext, &newClient)

				if errClientAdd != nil {
					fmt.Printf(errClientAdd.Error())

				} else {
					// Inform terminal
					fmt.Printf("Client connected: #%d from %s on port %d\n", newClient.socket, newClient.ip,
						newClient.port)

					// Inform connected client
					msg := "Welcome to Pong server!\n"
					_, _ = syscall.Write(newClient.socket, []byte(msg))
				}

			}
		}

		// Check for clients socket activity
		for _, client := range (*serverContext).clients {
			// Client activity
			if FD_ISSET(rfds, client.socket) {
				// Make 64 byte buffer
				buffer := make([]byte, 32)
				// Receive data from socket
				n, _, errRecv := syscall.Recvfrom(client.socket, buffer, 0)

				fmt.Printf("Client #%d received %d\n", client.socket, n)

				if n == 0 {
					_ = client.buffer.Flush()
					_ = client.writer.Close()
					_ = client.reader.Close()
					_ = RemoveClient(serverContext, client.socket)
					continue
				}

				if errRecv != nil {
					fmt.Printf("Receive error for client #%d: %s\n", client.socket, errRecv.Error())
				} else {
					// Write received data to client buffer
					_, errBuffW := client.buffer.Write(buffer)
					errFlushW := client.buffer.Flush()

					if errBuffW != nil {
						fmt.Printf("Buffer Write error for client #%d: %s\n", client.socket, errBuffW.Error())
					}

					if errFlushW != nil {
						fmt.Printf("Buffer Flush error for client #%d: %s\n", client.socket, errBuffW.Error())
					}
				}

			}
		}
	}
}

func Process(serverContext *server) {
	defer (*serverContext).WaitGroup.Done()

	fmt.Printf("Messages processing started!\n")
	defer fmt.Printf("Message proccessing terminated!\n")

	if serverContext == nil {
		msg := "Could not start server listener: server structure is NULL\n"
		fmt.Printf(msg)
		return
	}

	for {
		fmt.Printf("Processing tick\n")


		for _, currentClient := range (*serverContext).clients{
			go func(currentClient *client) {

				data := make([]byte, 32)
				n, errBuffRead := currentClient.buffer.Read(data)
				_ = currentClient.buffer.Flush()

				if errBuffRead != nil{
					return
				}

				if n < 1 {
					return
				}

				fmt.Printf("#%d> %v\n", currentClient.socket, data)
				msg := fmt.Sprintf("#%d> %s", currentClient.socket, string(data))
				_ = Broadcast(serverContext, []byte(msg))

			}(currentClient)
		}

		time.Sleep(100 * time.Millisecond)
	}

}
