// Integrate coatl with ambergris
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/bboreham/coatl/backends"
	"github.com/bboreham/coatl/data"
)

var backend *backends.Backend

func main() {
	backend = backends.NewBackend([]string{})
	run()
}

const SOCKET = "/var/run/ambergris.sock"

func sendToAmber(serviceName string) error {
	serviceInfo, err := backend.GetServiceDetails(serviceName)
	if err != nil {
		return err
	}
	var service data.Service
	if err := json.Unmarshal([]byte(serviceInfo), &service); err != nil {
		log.Println("Error unmarshalling: ", err)
		return err
	}
	var instances []data.Instance
	backend.ForeachInstance(serviceName, func(name, value string) {
		var instance data.Instance
		if err := json.Unmarshal([]byte(value), &instance); err != nil {
			log.Fatal("Error unmarshalling: ", err)
		}
		instances = append(instances, instance)
	})
	conn, err := net.Dial("unix", SOCKET)
	if err != nil {
		log.Println("Failed to connect to ambergris socket:", err)
		return err
	}
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("%s:%d", service.Address, service.Port))
	for _, instance := range instances {
		buf.WriteString(fmt.Sprintf(" %s:%d", instance.Address, instance.Port))
	}
	log.Println("Sending to ambergris:", buf.String())
	_, err = conn.Write(buf.Bytes())
	return err
}

func run() {
	ch := backend.Watch()

	for r := range ch {
		//fmt.Println(r.Action, r.Node)
		serviceName, _, err := data.DecodePath(r.Node.Key)
		if err != nil {
			log.Println(err)
			continue
		}
		if serviceName == "" {
			// everything deleted; can't cope
			continue
		}
		sendToAmber(serviceName)
	}
}
