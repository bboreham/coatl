package main

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/bboreham/coatl/backends"
	"github.com/bboreham/coatl/data"

	docker "github.com/fsouza/go-dockerclient"
)

type Listener struct {
	backend    *backends.Backend
	dc         *docker.Client
	services   map[string]*service
	containers map[string]*docker.Container
}

type service struct {
	name    string
	details data.Service
}

func NewListener(dc *docker.Client) *Listener {
	listener := &Listener{
		backend:    backends.NewBackend([]string{}),
		dc:         dc,
		services:   make(map[string]*service),
		containers: make(map[string]*docker.Container),
	}
	return listener
}

// Read in all info on registered services
func (l *Listener) ReadInServices() error {
	var s *service
	return l.backend.ForeachServiceInstance(func(name, value string) {
		s = &service{name: name}
		l.services[name] = s
		if err := json.Unmarshal([]byte(value), &s.details); err != nil {
			log.Println("Error unmarshalling:", err)
		}
	}, nil)
}

// Read details of all running containers
func (l *Listener) ReadExistingContainers() error {
	conts, err := l.dc.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return err
	}
	for _, cont := range conts {
		container, err := l.dc.InspectContainer(cont.ID)
		if err != nil {
			log.Println("Failed to inspect container:", cont.ID, err)
		}
		l.containers[cont.ID] = container
	}
	return nil
}

func (l *Listener) Sync() error {
	// Register all the ones we know about
	for _, container := range l.containers {
		l.Register(container)
	}
	// Remove all the ones we don't
	var serviceName string
	return l.backend.ForeachServiceInstance(func(name, value string) {
		serviceName = name
	}, func(instanceName, value string) {
		if _, found := l.containers[instanceName]; !found {
			log.Printf("Removing %.12s/%.12s", serviceName, instanceName)
			l.backend.RemoveInstance(serviceName, instanceName)
		}
	})
}

func (l *Listener) Register(container *docker.Container) error {
	service := l.serviceName(container)
	if err := l.backend.CheckRegisteredService(service); err != nil {
		log.Printf("ignoring %.12s; service '%s' not registered", container.ID, service)
		return nil
	}
	port := l.servicePort(container)
	if port == 0 {
		log.Printf("coatl: cannot find port for instance %.12s", container.ID)
		return nil
	}
	err := l.backend.AddInstance(service, container.ID, container.NetworkSettings.IPAddress, port)
	if err != nil {
		log.Println("coatl: failed to register service:", err)
		return err
	}
	log.Printf("Registered %s instance %.12s", service, container.ID)
	return nil
}

func (l *Listener) Deregister(container *docker.Container) error {
	service := l.serviceName(container)
	if l.backend.CheckRegisteredService(service) != nil {
		return nil
	}
	err := l.backend.RemoveInstance(service, container.ID)
	if err != nil {
		log.Println("coatl: failed to deregister service:", err)
	}
	log.Printf("Deregistered %s instance %.12s", service, container.ID)
	return err
}

func findOverride(container *docker.Container, key string) (val string, found bool) {
	for _, kv := range container.Config.Env {
		kvp := strings.SplitN(kv, "=", 2)
		if kvp[0] == key {
			return kvp[1], true
		}
	}
	// Todo: look in labels too
	return "", false
}

func (l *Listener) serviceName(container *docker.Container) string {
	// First choice is just the container name
	name := strings.TrimPrefix(container.Name, "/")
	// If there is an environment variable overriding, use that
	if val, found := findOverride(container, "SERVICE_NAME"); found {
		name = val
	}
	// If this is a service that has been registered against a specific image name, override
	for serviceName, service := range l.services {
		if container.Image == service.details.Image {
			name = serviceName
			break
		}
	}
	return name
}

func (l *Listener) servicePort(container *docker.Container) int {
	port := 0
	// If there is exactly one port exposed, that's the one.
	if len(container.NetworkSettings.Ports) == 1 {
		for portInfo := range container.NetworkSettings.Ports {
			if val, err := strconv.Atoi(portInfo.Port()); err == nil {
				port = val
			}
		}
	}
	// If there is an environment variable overriding, use that
	if val, found := findOverride(container, "SERVICE_PORT"); found {
		if num, err := strconv.Atoi(val); err != nil {
			port = num
		}
	}
	return port
}

func (l *Listener) Run(events <-chan *docker.APIEvents) {
	for event := range events {
		switch event.Status {
		case "start":
			container, err := l.dc.InspectContainer(event.ID)
			if err != nil {
				log.Println("Failed to inspect container:", event.ID, err)
			}
			l.containers[event.ID] = container
			l.Register(container)
		case "die":
			container, found := l.containers[event.ID]
			if !found {
				log.Println("Unknown container:", event.ID)
				continue
			}
			l.Deregister(container)
		}
	}
}
