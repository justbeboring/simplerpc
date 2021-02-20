/**
 * Copyright 2015-2017, Wothing Co., Ltd.
 * All rights reserved.
 *
 * Created by elvizlai on 2017/11/28 11:40.
 */

package registry

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"log"
	"net/http"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Consul struct {
	Cli *consulapi.Client
	Des string
	addr string
}

////init consul client
func (consul *Consul) Init(addr, des string) error {
	var err error
	consul.Cli, err = consulapi.NewClient(&consulapi.Config{Scheme: "http", Address: addr})
	if err != nil {
		return err
	}
	consul.addr = addr
	consul.Des = des
	return nil
}

// register service
func (consul *Consul) Register(name, addr string) error {
	var err error
	addr_elements := strings.Split(addr, ":")
	serviceIp := addr_elements[0]
	servicePort, _ := strconv.Atoi(addr_elements[1])
	checkaddr := fmt.Sprintf("%s:%d", serviceIp, servicePort+1)

	registration := &consulapi.AgentServiceRegistration{
		ID:      "/" + schema + "/" + name + fmt.Sprintf("/%s:%d", serviceIp, servicePort),
		Name:    "/" + schema + "/" + name,
		Port:    servicePort,
		Address: serviceIp,
		Tags:    []string{schema},
		Weights: &consulapi.AgentWeights{Passing: runtime.NumCPU()},
		Check: &consulapi.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s%s", checkaddr, "/check"),
			Timeout:                        "3s",
			Interval:                       "5s",
			DeregisterCriticalServiceAfter: "10s",
			//GRPC:fmt.Sprint("%v:%v/%v", service_ip, service_port,),
		},
	}

	err = consul.Cli.Agent().ServiceRegister(registration)
	if err != nil {
		log.Println(err)
	}

	////start check server
	http.HandleFunc("/check", consulCheck)
	go http.ListenAndServe(checkaddr, nil)

	log.Println(fmt.Sprintf("consul[%v] register success!",consul.addr))
	return nil
}

////http function to recieve check request
func consulCheck(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, "consulCheck")
}

// UnRegister remove service from etcd
func (consul *Consul) Unregister(name, addr string) error {
	return consul.Cli.Agent().ServiceDeregister(path.Join("/", schema, name, addr))
}

////get service by name
func (consul *Consul) GetService(name string) (map[string]int, error) {
	service := make(map[string]int)
	keyPrefix := "/" + schema + "/" + name
	getResp, err := consul.Cli.Agent().Services()
	if err != nil {
		log.Println(err)
	}
	for _, record := range getResp {
		addr := strings.TrimPrefix(string(record.ID), keyPrefix)
		weight, err := strconv.Atoi(string(record.Weights.Passing))
		if err != nil {
			service[addr] = 1
		} else {
			service[addr] = weight
		}
	}
	return service, nil
}

////list all service on this registry
func (consul *Consul) ListServices() (map[string]map[string]int, error) {
	keyPrefix := "/" + schema
	services := make(map[string]map[string]int)
	getResp, err := consul.Cli.Agent().Services()
	if err != nil {
		log.Println(err)
	}
	for _, record := range getResp {
		if strings.Index(record.ID, keyPrefix) == 0 {
			name := strings.TrimPrefix(string(record.ID), keyPrefix)
			if _, ok := services[name]; !ok {
				addr := strings.TrimPrefix(string(record.ID), keyPrefix)
				weight, err := strconv.Atoi(string(record.Weights.Passing))
				if err != nil {
					services[name][addr] = 1
				} else {
					services[name][addr] = weight
				}
			} else {
				service := make(map[string]int)
				services[name] = service
				addr := strings.TrimPrefix(string(record.ID), keyPrefix)
				weight, err := strconv.Atoi(string(record.Weights.Passing))
				if err != nil {
					service[addr] = 1
				} else {
					service[addr] = weight
				}
			}
		}
	}
	return services, nil
}

func (consul *Consul) Watch(keyPrefix string, events chan *Event) error {
	go func() {
		//var lastIndex uint64
		for {
			//getResp, metainfo, err := consul.Cli.Health().Service(keyPrefix, schema, true, &consulapi.QueryOptions{WaitIndex: lastIndex})
			//if err != nil {
			//	log.Println(err)
			//}
			//lastIndex = metainfo.LastIndex

			getResp, err := consul.Cli.Agent().ServicesWithFilter(`Service=="` + keyPrefix + `"`)
			if err != nil {
				log.Println(err)
			}

			for _, service := range getResp {
				//for i := 0; i < service.Weights.Passing; i++ {
					events <- &Event{0, fmt.Sprintf("%s:%d", service.Address, service.Port),1}
				//}
			}

			time.Sleep(time.Second)
		}
	}()

	return nil
}

////return registry
func (consul *Consul) String() string {
	return fmt.Sprintf("consul[%s]",consul.addr)
}
