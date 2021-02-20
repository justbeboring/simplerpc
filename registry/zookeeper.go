/**
 * Copyright 2015-2017, Wothing Co., Ltd.
 * All rights reserved.
 *
 * Created by elvizlai on 2017/11/28 11:40.
 */

package registry

import (
	"time"
	gozk "github.com/samuel/go-zookeeper/zk"
	"log"
	"bytes"
	"encoding/binary"
	"runtime"
	"fmt"
)

type Zk struct {
	Cli *gozk.Conn
	Des string
	addr string
}

func (zk *Zk)Init(addr string, des string)error{
	var err error
	zk.Des = des
	zk.addr = addr
	zk.Cli,_,err = gozk.Connect([]string{addr},time.Second)
	return err
}

// Register register service with name as prefix to etcd, multi etcd addr should use ; to split
func (zk *Zk)Register(name string, addr string) error {
	_,_,err := zk.Cli.Get("/"+schema)
	if err == gozk.ErrNoNode{
		_,err := zk.Cli.Create("/"+schema,nil,0,gozk.WorldACL(gozk.PermAll))
		if err != nil{
			log.Println(err)
		}
	}

	_,_,err = zk.Cli.Get("/"+schema+"/"+name)
	if err == gozk.ErrNoNode{
		_,err_create := zk.Cli.Create("/"+schema+"/"+name,nil,0,gozk.WorldACL(gozk.PermAll))
		if err_create != nil{
			log.Println(err_create)
		}
	}

	_,_,err = zk.Cli.Get("/"+schema+"/"+name+"/"+addr)
	if err == gozk.ErrNoNode{
		_,err_create := zk.Cli.Create("/"+schema+"/"+name+"/"+addr,int2Bytes(runtime.NumCPU()),gozk.FlagEphemeral,gozk.WorldACL(gozk.PermAll))
		if err != nil{
			log.Println(err_create)
		}
	}

	return nil
}

func (zk *Zk)Unregister(name,addr string)error {
	keyPrefix := "/" + schema + "/" + name + "/" + addr
	err := zk.Cli.Delete(keyPrefix,3)
	return err
}

func (zk *Zk)GetService(name string) (map[string]int, error){
	service := make(map[string]int)
	keyPrefix := name
	children,_,_ := zk.Cli.Children(keyPrefix)
	for _,child := range children{
		weight,_,_ := zk.Cli.Get(keyPrefix + "/" + child)
		count := int(weight[3])
		service[child] = count
	}
	return service,nil
}

func (zk *Zk) ListServices() (map[string]map[string]int, error) {
	services := make(map[string]map[string]int)
	keyPrefix := "/" + schema + "/"
	children, _, _ := zk.Cli.Children(keyPrefix)
	for _, child := range children {
		service, _ := zk.GetService(child)
		services[child] = service
	}
	return services, nil
}

func (zk *Zk) Watch(keyPrefix string, events chan *Event) error {
	if zk.Cli == nil {
		return gozk.ErrConnectionClosed
	}
	//_, _, msgChan, _ := zk.Cli.ExistsW(keyPrefix + "/127.0.0.1:50053")
	//
	//go func() {
	//	for {
	//		msg := <-msgChan
	//		if msg.Type == 4 {
	//			events <- &registry.Event{int(msg.Type), msg.Path}
	//		}
	//	}
	//}()

	go func() {
		for {
			services, _ := zk.GetService(keyPrefix)
			for addr, _ := range services {
				events <- &Event{0, addr,1}
			}
			time.Sleep(time.Second)
		}
	}()

	return nil
}

func (zk *Zk)String() string{
	return fmt.Sprintf("zookeeper[%s]",zk.addr)
}

func int2Bytes(n int)[]byte{
	tmp := uint32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer,binary.BigEndian,&tmp)
	return bytesBuffer.Bytes()
}
