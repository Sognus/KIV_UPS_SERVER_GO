package main

import (
	"./communication"
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
		fmt.Printf(errInit.Error())
		os.Exit(-1)
	}

	// Run goroutines
	(*serverContext).WaitGroup.Add(1)
	go communication.Start(serverContext)

	// Wait for all goroutines to end
	(*serverContext).WaitGroup.Wait()

}
