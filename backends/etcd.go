package backends

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/bboreham/coatl/data"
	"github.com/coreos/go-etcd/etcd"
)

type Backend struct {
	client *etcd.Client
}

func NewBackend(machines []string) *Backend {
	if len(machines) == 0 {
		machines = []string{"http://127.0.0.1:4001"}
	}
	backend := &Backend{client: etcd.NewClient(machines)}
	return backend
}

func (b *Backend) GetService(serviceName string) error {
	_, err := b.client.Get(data.ServicePath+serviceName, false, false)
	return err
}

func (b *Backend) AddService(serviceName, address string, port int, image string) {
	if _, err := b.client.CreateDir(data.ServicePath+serviceName, 0); err != nil {
		log.Fatal("Unable to write:", err)
	}
	details := data.Service{Address: address, Port: port, Image: image}
	json, err := json.Marshal(&details)
	if err != nil {
		log.Fatal("Failed to encode:", err)
	}
	b.client.Set(data.ServicePath+serviceName+"/_details", string(json), 0)
}

func (b *Backend) RemoveService(serviceName string) error {
	_, err := b.client.Delete(data.ServicePath+serviceName, true)
	return err
}

func (b *Backend) RemoveAllServices() error {
	_, err := b.client.Delete(data.ServicePath, true)
	return err
}

func (b *Backend) ForeachServiceInstance(fs, fi func(string, string)) {
	r, err := b.client.Get(data.ServicePath, true, fi != nil)
	if err != nil {
		if etcderr, ok := err.(*etcd.EtcdError); ok && etcderr.ErrorCode == 100 {
			return
		}
		log.Fatal("Failed to get data:", err)
	}
	for _, node := range r.Node.Nodes {
		details, err := b.client.Get(node.Key+"/_details", false, false)
		if err != nil {
			log.Fatal("Failed to get data:", err)
		}
		fs(strings.TrimPrefix(node.Key, data.ServicePath), details.Node.Value)
		for _, instance := range node.Nodes {
			fi(strings.TrimPrefix(instance.Key, node.Key), instance.Value)
		}
	}
}

func (b *Backend) AddInstance(serviceName, instanceName, address string, port int) {
	details := data.Instance{Address: address, Port: port}
	json, err := json.Marshal(&details)
	if err != nil {
		log.Fatal("Failed to encode:", err)
	}
	if _, err := b.client.Set(data.ServicePath+serviceName+"/"+instanceName, string(json), 0); err != nil {
		log.Fatal("Unable to write:", err)
	}
}

func (b *Backend) RemoveInstance(serviceName, instanceName string) {
	_, err := b.client.Delete(data.ServicePath+serviceName+"/"+instanceName, true)
	if err != nil {
		log.Fatal("Failed to delete:", err)
	}
}

// Needs work to make less etcd-centric
func (b *Backend) Watch() chan *etcd.Response {
	ch := make(chan *etcd.Response)
	go b.client.Watch(data.ServicePath, 0, true, ch, nil)
	return ch
}
