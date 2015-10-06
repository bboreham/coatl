package main

import (
	"encoding/json"
	"log"

	"github.com/bboreham/coatl/backends"
	"github.com/bboreham/coatl/data"

	docker "github.com/fsouza/go-dockerclient"
)

const (
	servicePath = "/weave/service/"
	dockerPath  = "unix:///var/run/docker.sock"
)

var (
	dc *docker.Client
)

func setupDockerClient(apiPath string) (*docker.Client, error) {
	dc, err := docker.NewClient(apiPath)
	if err != nil {
		return nil, err
	}
	env, err := dc.Version()
	if err != nil {
		return nil, err
	}
	log.Printf("[docker] Using Docker API on %s: %v", apiPath, env)
	return dc, nil
}

type service struct {
	name    string
	details data.Service
}

var services map[string]*service

func createService(name string) *service {
	s := &service{name: name}
	services[name] = s
	return s
}

// Read in all info on services
func initialize() {
	services = make(map[string]*service)
	var s *service
	backend.ForeachServiceInstance(func(name, value string) {
		s = createService(name)
		if err := json.Unmarshal([]byte(value), &s.details); err != nil {
			log.Fatal("Error unmarshalling: ", err)
		}
	}, nil)
}

var backend *backends.Backend

func main() {
	backend = backends.NewBackend([]string{})
	initialize()
	dc, err := setupDockerClient(dockerPath)
	if err != nil {
		log.Fatal(err)
	}

	events := make(chan *docker.APIEvents)
	if err := dc.AddEventListener(events); err != nil {
		log.Fatalf("[docker] Unable to add listener to Docker API: %s", err)
	}

	go func() {
		for event := range events {
			switch event.Status {
			case "start":
				container, err := dc.InspectContainer(event.ID)
				if err != nil {
					log.Fatal("Failed to inspect container:", event.ID, err)
				}
				backend.AddInstance("foo", container.Name, container.NetworkSettings.IPAddress, 1234)
			case "die":
				container, err := dc.InspectContainer(event.ID)
				if err != nil {
					log.Fatal("Failed to inspect container:", event.ID, err)
				}
				backend.RemoveInstance("foo", container.Name)
			}
		}
	}()

}
