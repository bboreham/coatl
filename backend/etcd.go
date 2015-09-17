package backend

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/bboreham/coatl/data"
	"github.com/coreos/go-etcd/etcd"
)

// urgh singleton
var client *etcd.Client

func SetupBackend() {
	machines := []string{"http://127.0.0.1:4001"}
	client = etcd.NewClient(machines)
}

func GetService(serviceName string) error {
	_, err := client.Get(data.ServicePath+serviceName, false, false)
	return err
}

func AddService(serviceName, address string, port int) {
	if _, err := client.CreateDir(data.ServicePath+serviceName, 0); err != nil {
		log.Fatal("Unable to write:", err)
	}
	details := data.Service{Address: address, Port: port}
	json, err := json.Marshal(&details)
	if err != nil {
		log.Fatal("Failed to encode:", err)
	}
	client.Set(data.ServicePath+serviceName+"/_details", string(json), 0)
}

func RemoveService(serviceName string) error {
	_, err := client.Delete(data.ServicePath+serviceName, true)
	return err
}

func RemoveAllServices() error {
	_, err := client.Delete(data.ServicePath, true)
	return err
}

func ForeachServiceInstance(all bool, fs, fi func(string, string)) {
	r, err := client.Get(data.ServicePath, true, all)
	if err != nil {
		if etcderr, ok := err.(*etcd.EtcdError); ok && etcderr.ErrorCode == 100 {
			return
		}
		log.Fatal("Failed to get data:", err)
	}
	for _, node := range r.Node.Nodes {
		details, err := client.Get(node.Key+"/_details", false, false)
		if err != nil {
			log.Fatal("Failed to get data:", err)
		}
		fs(strings.TrimPrefix(node.Key, data.ServicePath), details.Node.Value)
		for _, instance := range node.Nodes {
			fi(strings.TrimPrefix(instance.Key, node.Key), instance.Value)
		}
	}
}

func AddInstance(serviceName, instanceName, address string, port int) {
	details := data.Instance{Address: address, Port: port}
	json, err := json.Marshal(&details)
	if err != nil {
		log.Fatal("Failed to encode:", err)
	}
	if _, err := client.Set(data.ServicePath+serviceName+"/"+instanceName, string(json), 0); err != nil {
		log.Fatal("Unable to write:", err)
	}
}

func RemoveInstance(serviceName, instanceName string) {
	_, err := client.Delete(data.ServicePath+serviceName+"/"+instanceName, true)
	if err != nil {
		log.Fatal("Failed to delete:", err)
	}
}

// Needs work to make less etcd-centric
func Watch() chan *etcd.Response {
	ch := make(chan *etcd.Response)
	go client.Watch(data.ServicePath, 0, true, ch, nil)
	return ch
}
