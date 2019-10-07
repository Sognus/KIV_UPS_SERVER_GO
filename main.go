package main

import (
	"fmt"
	"os"
	"syscall"
)

func main() {
	fmt.Println("Starting server")

	main_socket, socket_err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)

	if socket_err != nil {
		fmt.Println("Could not create main socket!")
		os.Exit(1)
	}
	fmt.Println("Created main socket!")

	address := syscall.SockaddrInet4{
		Port: 8081,
		Addr: [4]byte{192, 168, 0, 109},
	}

	bind_err := syscall.Bind(main_socket, &address)

	if bind_err != nil {
		fmt.Println("Could not bind given IP adress!")
		os.Exit(2)
	}
	fmt.Printf("IP adress %d.%d.%d.%d was bound!\n", address.Addr[0], address.Addr[1], address.Addr[2], address.Addr[3])

	listen_err := syscall.Listen(main_socket, 10);

	if listen_err != nil {
		fmt.Println("Could not start listener for given IP adress!")
		os.Exit(3)
	}
	println("Listener was started!\n")

	for {
		println("Accepting new connection...")
		new_socket, new_socket_info , accept_err := syscall.Accept(main_socket)

		if accept_err != nil {
			println("Error occured while accepting new connection: ", accept_err)
		}

		msg := []byte("Nazdar, jsi připojen k serveru!");

		fmt.Println("Client socket file descriptor: ", new_socket)
		fmt.Println("Client port: ", new_socket_info)

		n, write_err := syscall.Write(new_socket, msg)

		if write_err != nil {
			println("Could not send message to client!")
		}

		if n != len(msg) {
			println("Message was sent but didnt arrive whole!")
		}


	}
}
