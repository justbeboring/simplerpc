/**
 * Copyright 2015-2017, Wothing Co., Ltd.
 * All rights reserved.
 *
 * Created by elvizlai on 2017/11/28 11:40.
 */

package registry

import (
	"time"
	goredis "github.com/go-redis/redis"
	"fmt"
	"runtime"
	"errors"
	"strings"
	"strconv"
)

type Redis struct {
	Cli *goredis.Client
	Des string
	addr string
}

func (redis *Redis) Init(addr, des string) error {
	redis.Cli = goredis.NewClient(&goredis.Options{
		Network:  "tcp",
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	redis.Des = des
	redis.addr = addr
	return nil
}

// Register register service with name as prefix to etcd, multi etcd addr should use ; to split
func (redis *Redis) Register(name, addr string) error {
	if redis.Cli == nil {
		return errors.New("registry client is nil")
	}

	ticker := time.NewTicker(time.Second * time.Duration(1))

	go func() {
		for {
			redis.Cli.Publish("/"+schema+"/"+name, "keep\t"+addr+fmt.Sprintf("\t%d", runtime.NumCPU()))
			<-ticker.C
		}
	}()

	return nil
}

// UnRegister remove service from etcd
func (redis *Redis) Unregister(name, addr string) error {
	if redis.Cli != nil {
		return errors.New("registry client is nil")
	}
	redis.Cli.Publish("/"+schema+"/"+name, "del\t"+addr)
	return nil
}

func (redis *Redis) GetService(name string) (map[string]int, error) {
	service := make(map[string]int)
	keyPrefix := "/" + schema + "/" + name
	if redis.Cli != nil {
		return service, errors.New("registry client is nil")
	}
	msgChan := redis.Cli.Subscribe(keyPrefix).Channel()
	time.Sleep(time.Second * 2)
	for msg := range msgChan {
		msgStr := strings.Split(msg.Payload, "\t")
		switch msgStr[0] {
		case "keep":
			weight, _ := strconv.Atoi(msgStr[2])
			service[msgStr[1]] = weight
		default:

		}
	}
	return service, nil
}
func (redis *Redis) ListServices() (map[string]map[string]int, error) {
	services := make(map[string]map[string]int)
	keyPrefix := "/" + schema
	if redis.Cli != nil {
		return services, errors.New("registry client is nil")
	}
	getResp := redis.Cli.PubSubChannels(keyPrefix)
	msgChanMap := make(map[string]<-chan *goredis.Message)
	for _, val := range getResp.Val() {
		msgChanMap[val] = redis.Cli.Subscribe(keyPrefix).Channel()
	}
	return services, nil
}

func (redis *Redis) Watch(keyPrefix string, events chan *Event) error {
	msgChan := redis.Cli.Subscribe(keyPrefix).Channel()

	go func() {
		for {
			msg := <-msgChan
			msgStr := strings.Split(msg.Payload, "\t")
			weight,_ := strconv.Atoi(msgStr[2])
			if msgStr[0] == "keep" {
				events <- &Event{0, msgStr[1],weight}
			} else {
				events <- &Event{1, msgStr[1],weight}
			}
			time.Sleep(time.Second)
		}
	}()

	return nil
}

func (redis *Redis) String() string {
	return fmt.Sprintf("redis[%s]",redis.addr)
}
