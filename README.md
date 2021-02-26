# simplerpc
a simple rpc-frame over grpc with registry(etcd/consul/zookeeper/redis) and load-balance(weighted-round-roubin/weighted-random).

server：

	service := simplerpc.NewService("test",addr)
	service.AddRegistry("etcd","127.0.0.1:2379","")
	//service.AddRegistry("conaul","11.36.208.249:8500","")
	//service.AddRegistry("zookeeper","11.36.208.249:2181","")
	//service.AddRegistry("redis","11.36.208.249:6379","")
	//service.SetCreds("tls/server.crt","tls/server.key")
	service.Init()
	pb.RegisterHelloServiceServer(service.GrpcServer, &hello{})
	service.Run()

client：

	r := simplerpc.NewResolver("test")
	r.AddRegistry("etcd","127.0.0.1:2379","")
	//r.AddRegistry("cansul","11.36.208.249:8500","")
	//r.AddRegistry("zookeeper","11.36.208.249:2181","")
	//r.AddRegistry("redis", "11.36.208.249:6379", "")
	//r.SetCreds("tls/server.crt","server.grpc.io")
	resolver.Register(r)
	r.Init()
	client := pb.NewHelloServiceClient(r.Conn)
	resp, err := client.Echo(context.Background(), &pb.Payload{Data: "hello"}, grpc.FailFast(true))
	if err != nil {
		log.Println(err)
	} else {
		log.Println(resp.Data)
	}
