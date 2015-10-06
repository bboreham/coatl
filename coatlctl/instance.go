package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/bboreham/coatl/backend"
)

func enrol(args []string) {
	if len(args) != 4 {
		log.Fatal("Usage: coatlctl enrol <service> <instance> <address> <port>")
	}
	serviceName, instance := args[0], args[1]
	if err := backend.GetService(serviceName); err != nil {
		log.Fatal("Cannot find service '", serviceName, "':", err)
	}
	port, err := strconv.Atoi(args[3])
	if err != nil {
		log.Fatal("Invalid port number: ", err)
	}
	backend.AddInstance(serviceName, instance, args[2], port)
	fmt.Println("Enrolled", instance, "in service", serviceName)
}

func unenrol(args []string) {
	if len(args) != 2 {
		log.Fatal("Usage: coatlctl unenrol <service> <instance>")
	}
	serviceName, instance := args[0], args[1]
	backend.RemoveInstance(serviceName, instance)
	fmt.Println("Un-enrolled", instance, "from service", serviceName)
}