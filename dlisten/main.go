package main

import (
	"log"

	docker "github.com/fsouza/go-dockerclient"
)

const (
	dockerPath = "unix:///var/run/docker.sock"
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

func main() {
	dc, err := setupDockerClient(dockerPath)
	if err != nil {
		log.Fatal(err)
	}
	listener := NewListener(dc)

	events := make(chan *docker.APIEvents)
	if err := dc.AddEventListener(events); err != nil {
		log.Fatalf("[docker] Unable to add listener to Docker API: %s", err)
	}

	listener.Run(events)
}
