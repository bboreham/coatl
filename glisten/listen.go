// Integrate coatl with ambergris via channel
package glisten

import (
	"encoding/json"
	"log"
	"net"

	"github.com/bboreham/coatl/backends"
	"github.com/bboreham/coatl/data"
	"github.com/dpw/ambergris/interceptor/model"
)

type Listener struct {
	backend *backends.Backend
	updates chan model.ServiceUpdate
	errors  chan<- error
}

func (l *Listener) send(serviceName string) error {
	serviceInfo, err := l.backend.GetServiceDetails(serviceName)
	if err != nil {
		return err
	}
	var service data.Service
	if err := json.Unmarshal([]byte(serviceInfo), &service); err != nil {
		log.Println("Error unmarshalling: ", err)
		return err
	}
	update := model.ServiceUpdate{
		ServiceKey:  model.MakeServiceKey("tcp", net.ParseIP(service.Address), service.Port),
		ServiceInfo: &model.ServiceInfo{},
	}
	l.backend.ForeachInstance(serviceName, func(name, value string) {
		var instance data.Instance
		if err := json.Unmarshal([]byte(value), &instance); err != nil {
			log.Fatal("Error unmarshalling: ", err)
		}
		update.ServiceInfo.Instances = append(update.ServiceInfo.Instances, model.MakeInstance(net.ParseIP(instance.Address), instance.Port))
	})
	log.Printf("Sending update for %s: %+v\n", update.ServiceKey.String(), update.ServiceInfo)
	l.updates <- update
	return nil
}

func NewListener(errors chan<- error) (*Listener, error) {
	listener := &Listener{
		backend: backends.NewBackend([]string{}),
		updates: make(chan model.ServiceUpdate),
		errors:  errors,
	}
	go listener.run()
	return listener, nil
}

func (l *Listener) Updates() <-chan model.ServiceUpdate {
	return l.updates
}

func (l *Listener) run() {
	ch := l.backend.Watch()

	for r := range ch {
		log.Println(r.Action, r.Node)
		serviceName, _, err := data.DecodePath(r.Node.Key)
		if err != nil {
			log.Println(err)
			continue
		}
		if serviceName == "" {
			// everything deleted; can't cope
			continue
		}
		l.send(serviceName)
	}
}

func (l *Listener) Close() {
	// TODO
}
