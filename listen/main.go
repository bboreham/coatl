package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bboreham/coatl/backend"
	"github.com/bboreham/coatl/data"
	"github.com/spf13/cobra"
)

var topCmd = &cobra.Command{
	Use:   "listen",
	Short: "listen to weave Run updates",
	Long:  `Write more documentation here`,
	Run:   run,
}

func main() {
	backend.SetupBackend()

	err := topCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

const (
	servicePath = "/weave/service/"
)

type instance struct {
	name    string
	details data.Instance
}

type service struct {
	name      string
	details   data.Service
	instances map[string]*instance
}

var services map[string]*service

func createService(name string) *service {
	s := &service{name: name, instances: make(map[string]*instance)}
	services[name] = s
	return s
}

func createInstance(s *service, name string, data string) *instance {
	i := &instance{name: name}
	if err := json.Unmarshal([]byte(data), &i.details); err != nil {
		log.Fatal("Error unmarshalling: ", err)
	}
	s.instances[i.name] = i
	return i
}

func initialize() {
	services = make(map[string]*service)
	var s *service
	backend.ForeachServiceInstance(true, func(name, value string) {
		s = createService(name)
		if err := json.Unmarshal([]byte(value), &s.details); err != nil {
			log.Fatal("Error unmarshalling: ", err)
		}
	}, func(name, value string) {
		createInstance(s, name, value)
	})
}

func run(cmd *cobra.Command, args []string) {
	initialize()
	fmt.Print(len(services), " services:")
	for name := range services {
		fmt.Print(" ", name)
	}
	fmt.Println()
	ch := backend.Watch()

	for r := range ch {
		//fmt.Println(r.Action, r.Node)
		serviceName, instanceName, err := data.DecodePath(r.Node.Key)
		if err != nil {
			log.Println(err)
			continue
		}
		switch r.Action {
		case "create":
			createService(serviceName)
			fmt.Println("Service created:", serviceName, "; there are now", len(services), "services")
		case "delete":
			if serviceName == "" {
				// everything deleted
				services = make(map[string]*service)
				fmt.Println("All services deleted")
			} else if instanceName == "" {
				delete(services, serviceName)
				fmt.Println("Service deleted:", serviceName, "; there are now", len(services), "services")
			} else {
				s, ok := services[serviceName]
				if !ok {
					log.Println("Service not found:", serviceName)
					continue
				}
				delete(s.instances, instanceName)
				fmt.Println("Instance", instanceName, "removed from", s.name, "which now has", len(s.instances), "instances")
			}
		case "set":
			s, ok := services[serviceName]
			if !ok {
				log.Println("Service not found:", serviceName)
				continue
			}
			if instanceName == "_details" {
				if err := json.Unmarshal([]byte(r.Node.Value), &s.details); err != nil {
					log.Println("Error unmarshalling: ", err)
					continue
				}
				fmt.Println("Service", s.name, s.details)
			} else {
				i := createInstance(s, instanceName, r.Node.Value)
				fmt.Println("Instance", i.name, "is now enrolled in", s.name, "which now has", len(s.instances), "instances")
			}
		default:
			fmt.Println("Unhandled:", r.Action, r.Node)
		}
	}
}
