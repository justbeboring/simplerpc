package main

import (
	"context"
	"github.com/justbeboring/simplerpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"log"
	"main/pb"
)

func main() {
	r := simplerpc.NewResolver("test")

	r.AddRegistry("etcd","127.0.0.1:2379","")
	//r.AddRegistry("consul","127.0.0.1:8500","")
	//r.AddRegistry("zookeeper","127.0.0.1:2181","")
	//r.AddRegistry("redis","127.0.0.1:6379","")
	//r.SetCreds("tls/server.crt","server.grpc.io")

	resolver.Register(r)
	var err error
	r.Init()
	if err != nil {
		panic(err)
	}

	client := pb.NewHelloServiceClient(r.Conn)

	for i :=0;i< 20;i++{
		resp, err := client.Echo(context.Background(), &pb.Payload{Data: "hello"}, grpc.FailFast(true))
		if err != nil {
			log.Println(err)
		} else {
			log.Println(resp.Data)
		}
	}

}
