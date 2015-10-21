package data

import (
	"fmt"
	"strings"
)

type Service struct {
	Address  string `json:"address,omitempty"`
	Port     int    `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Image    string `json:"image,omitempty"`
}

type Instance struct {
	Address string `json:"address,omitempty"`
	Port    int    `json:"port,omitempty"`
}

const ServicePath = "/weave/service/"

func DecodePath(path string) (serviceName, instanceName string, err error) {
	if path+"/" == ServicePath {
		return "", "", nil
	}
	part := strings.Split(path, "/")
	if len(part) < 4 {
		return "", "", fmt.Errorf("bad path: %s", path)
	} else if len(part) < 5 {
		return part[3], "", nil
	}
	return part[3], part[4], nil
}
