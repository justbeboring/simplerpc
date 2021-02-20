module main

go 1.15

require (
	github.com/golang/protobuf v1.4.3
	github.com/justbeboring/simplerpc v0.0.0-20210220011107-5c9c5350854a
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	google.golang.org/grpc v1.35.0
)

replace (
	github.com/coreos/bbolt v1.3.5 => go.etcd.io/bbolt v1.3.5
	google.golang.org/grpc v1.35.0 => google.golang.org/grpc v1.26.0
)
