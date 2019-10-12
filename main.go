package main

import (
	"fmt"
	"os"
	"syscall"
)

func FD_SET(p *syscall.FdSet, i int) {
	p.Bits[i/64] |= 1 << uint(i) % 64
}

func FD_ISSET(p *syscall.FdSet, i int) bool {
	return (p.Bits[i/64] & (1 << uint(i) % 64)) != 0
}

func FD_ZERO(p *syscall.FdSet) {
	for i := range p.Bits {
		p.Bits[i] = 0
	}
}

type Client struct {
	socketDescriptor int
	address          syscall.Sockaddr
}

func main() {
	const PORT = 8080
	var ADDRESS = [4]byte{192, 168, 0, 108}

	fmt.Println("Starting server")

	main_socket, socket_err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)

	if socket_err != nil {
		fmt.Println("Could not create main socket!")
		os.Exit(1)
	}
	fmt.Println("Created main socket!")

	address := syscall.SockaddrInet4{
		Port: PORT,
		Addr: ADDRESS,
	}

	bind_err := syscall.Bind(main_socket, &address)

	if bind_err != nil {
		fmt.Println("Could not bind given IP adress!")
		os.Exit(2)
	}
	fmt.Printf("IP adress %d.%d.%d.%d:%d was bound!\n", address.Addr[0], address.Addr[1], address.Addr[2], address.Addr[3], address.Port)

	listen_err := syscall.Listen(main_socket, 10);
	if listen_err != nil {
		fmt.Println("Could not start listener for given IP adress!")
		os.Exit(3)
	}
	println("Listener was started!\n")

	// File descriptor set
	rfds := &syscall.FdSet{}
	timeout := &syscall.Timeval{}
	clients := make([]Client, 0, 10)

	// Print information
	println("Accepting new clients...")

	// Main communication Loop
	for {
		// Reset FD SET
		FD_ZERO(rfds)

		// Add main socket to socket set
		FD_SET(rfds, main_socket)

		max_sd := -1

		// Add established clients to socket set
		for i := 0; i < len(clients); i++ {
			socketFileDescriptor := clients[i].socketDescriptor
			FD_SET(rfds, socketFileDescriptor)
			// Find out highest socket file descriptor
			if socketFileDescriptor > max_sd {
				max_sd = socketFileDescriptor
			}
		}

		// TODO: find out what first parameter means
		_, errSelect := syscall.Select(max_sd+1, rfds, nil, nil, timeout)

		// Check if we got error
		if errSelect != nil {
			fmt.Printf("Select error: %s\n", errSelect.Error())
		}

		// Check communication for main socket
		if FD_ISSET(rfds, main_socket) {
			newSocketDescriptor, newAddress, errAccept := syscall.Accept(main_socket)

			// Check for accept error
			if errAccept != nil {
				fmt.Printf("Accept error: %s\n", errAccept.Error())
			}

			// Create new client record
			newClient := Client{
				socketDescriptor: newSocketDescriptor,
				address:          newAddress,
			}

			// Print information about new client
			address := newAddress.(*syscall.SockaddrInet4)
			fmt.Printf("New client accepted: %d.%d.%d.%d:%d\n", address.Addr[0], address.Addr[1], address.Addr[2], address.Addr[3], address.Port)

			// Sent greetings message to new client
			msg := "Připojen k serveru\n"
			_, errWrite := syscall.Write(newSocketDescriptor, []byte(msg))

			// Check for errors
			if errWrite != nil {
				fmt.Printf("Greet message write error for %d: %s\n", newSocketDescriptor, errWrite.Error())
			}

			// Save new client in storage
			clients = append(clients, newClient)
		}

		// Check for communication from other clients
		for i := 0; i < len(clients); i++ {
			clientReadSocketDescriptor := clients[i].socketDescriptor

			if FD_ISSET(rfds, clientReadSocketDescriptor) {
				buffer := make([]byte, 10)
				_, _, errRecv := syscall.Recvfrom(clientReadSocketDescriptor, buffer, 0)

				// Check receive error
				if errRecv != nil {
					fmt.Printf("Receive error: %s\n", errSelect.Error())
				}

				// Send to other clients
				for y := 0; y < len(clients); y++ {
					clientWriteSocketDescriptor := clients[y].socketDescriptor
					msg := fmt.Sprintf("<Client %d> %s\n", clientReadSocketDescriptor, buffer)
					// Todo: find out what is first value
					_, errWrite := syscall.Write(clientWriteSocketDescriptor, []byte(msg))

					if errWrite != nil {
						fmt.Printf("Write error for %d: %s\n", clientWriteSocketDescriptor, errWrite.Error())
					}

				}

			}

		}

	}
}
