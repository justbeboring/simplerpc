/**
 * Copyright 2015-2017, Wothing Co., Ltd.
 * All rights reserved.
 *
 * Created by elvizlai on 2017/11/28 11:40.
 */

package registry

import (
	"context"
	"log"
	"strings"
	"time"
	"github.com/coreos/etcd/clientv3"
	"runtime"
	"strconv"
	"path"
	"fmt"
)

type Etcd struct {
	Cli *clientv3.Client
	Des string
	addr string
}

func (etcd *Etcd) Init(addr string, des string) error {
	var err error
	etcd.Cli, err = clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(addr, ";"),
		DialTimeout: 15 * time.Second,
	})

	etcd.Des = des
	etcd.addr = addr
	return err
}

// Register register service with name as prefix to etcd, multi etcd addr should use ; to split
func (etcd *Etcd) Register(name, addr string) error {
	ticker := time.NewTicker(time.Second * time.Duration(3))
	if etcd.Cli != nil {
		go func() {
			for {
				getResp, err := etcd.Cli.Get(context.Background(), "/"+schema+"/"+name+"/"+addr)
				if err != nil {
					log.Println(err)
				} else if getResp.Count == 0 {
					err = etcd.withAlive(name, addr, 3)
					if err != nil {
						log.Println(err)
					}
				} else {
					// do nothing
				}
				<-ticker.C
			}
		}()
	} else {
		return clientv3.ErrNoAvailableEndpoints
	}
	log.Println(fmt.Sprintf("etcd[%v] register success!",etcd.addr))
	return nil
}

// UnRegister remove service from etcd
func (etcd *Etcd) Unregister(name, addr string) error {
	keyPrefix := path.Join("/", schema, name)
	if etcd.Cli != nil {
		_, err := etcd.Cli.Delete(context.Background(), keyPrefix+"/"+addr)
		if err != nil {
			return err
		}
	} else {
		return clientv3.ErrNoAvailableEndpoints
	}
	return nil
}

func (etcd *Etcd) GetService(name string) (map[string]int, error) {
	service := make(map[string]int)
	if etcd.Cli == nil {
		return service, clientv3.ErrNoAvailableEndpoints
	}
	getResp, err := etcd.Cli.Get(context.Background(), name, clientv3.WithPrefix())
	if err != nil {
		log.Println(err)
		return service, nil
	} else {
		for i := range getResp.Kvs {
			addr := strings.TrimPrefix(string(getResp.Kvs[i].Key), name + "/")
			weight, err := strconv.Atoi(string(getResp.Kvs[i].Value))
			if err != nil {
				service[addr] = weight
			} else {
				service[addr] = 1
			}
		}
	}
	return service, nil
}

func (etcd *Etcd) ListServices() (map[string]map[string]int, error) {
	services := make(map[string]map[string]int)
	keyPrefix := "/" + schema
	if etcd.Cli == nil {
		return services, clientv3.ErrNoAvailableEndpoints
	}
	getResp, err := etcd.Cli.Get(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		log.Println(err)
		return services, nil
	} else {
		for i := range getResp.Kvs {
			name := strings.TrimPrefix(string(getResp.Kvs[i].Key), keyPrefix)
			if _, ok := services[name]; !ok {
				addr := strings.TrimPrefix(string(getResp.Kvs[i].Key), keyPrefix)
				weight, err := strconv.Atoi(string(getResp.Kvs[i].Value))
				if err != nil {
					services[name][addr] = 1
				} else {
					services[name][addr] = weight
				}
			} else {
				service := make(map[string]int)
				services[name] = service
				addr := strings.TrimPrefix(string(getResp.Kvs[i].Key), keyPrefix)
				weight, err := strconv.Atoi(string(getResp.Kvs[i].Value))
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

func (etcd *Etcd) Watch(keyPrefix string, events chan *Event) error {
	if etcd.Cli == nil {
		return clientv3.ErrNoAvailableEndpoints
	}
	msgChan := etcd.Cli.Watch(context.Background(), keyPrefix, clientv3.WithPrefix())

	services,_ := etcd.GetService(keyPrefix)
	for addr,_ := range services{
		events <- &Event{0, addr,1}
	}

	go func() {
		for {
			msg := <-msgChan
			for _, ev := range msg.Events {
				addr := strings.TrimPrefix(string(ev.Kv.Key), keyPrefix + "/")
				events <- &Event{int(ev.Type), addr,1}
			}
		}
	}()

	return nil
}

func (etcd *Etcd) String() string {
	return fmt.Sprintf("etcd[%s]",etcd.addr)
}

func (etcd *Etcd) withAlive(name string, addr string, ttl int64) error {
	leaseResp, err := etcd.Cli.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}

	_, err = etcd.Cli.Put(context.Background(), "/"+schema+"/"+name+"/"+addr, strconv.Itoa(runtime.NumCPU()), clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	_, err = etcd.Cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}
	return nil
}
