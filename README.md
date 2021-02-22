# simplerpc
a simple rpc-frame over grpc with registry(etcd/consul/zookeeper/redis) and load-balance(weighted-round-roubin/weighted-random).

server：

    listener, err := net.Listen("tcp", "127.0.0.1:32449")
	server := grpc.NewServer()

	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	pb.RegisterHelloServiceServer(server,&hello{})
	server.Serve(listener)

client：

    conn, err := grpc.Dial(127.0.0.1:32449",grpc.WithInsecure())
	if err != nil {
		log.Print(err)
	}
	defer conn.Close()
	client := pb.NewHelloServiceClient(conn)
	for {
		resp, _ := client.Echo(context.Background(), &pb.Payload{Data: "hello"}, grpc.FailFast(true))
		fmt.Println(resp.Data)
	}
