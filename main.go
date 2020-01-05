package main

import (
	"./communication"
	"./game"
	"fmt"
	"os"
)

func main() {
	// Check if we have enough arguments to start communication
	if len(os.Args) < 3 {
		fmt.Printf("Missing arguments - atleast 2 needed, %d given\n", len(os.Args)-1)
		fmt.Printf("Usage: ./communication <ip> <port>")
		os.Exit(0)
	}

	// Initialize communication
	serverContext, errInit := communication.Init(os.Args[1], os.Args[2])

	if errInit != nil {
		fmt.Println(errInit.Error())
		os.Exit(-1)
	}

	// Initialize server manager
	serverManager, errManagerInit := game.ManagerInitialize(serverContext)

	if errManagerInit != nil {
		fmt.Println(errManagerInit.Error())
		os.Exit(-2)
	}

	// Run goroutines
	(*serverContext).WaitGroup.Add(1)
	go game.ManagerStart(serverContext, serverManager)
	(*serverContext).WaitGroup.Add(1)
	go communication.Start(serverContext)

	// Wait for all goroutines to end
	(*serverContext).WaitGroup.Wait()

}
