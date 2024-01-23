package main

import (
	"log"
	"net"

	astriaGrpc "buf.build/gen/go/astria/astria/grpc/go/astria/execution/v1alpha2/executionv1alpha2grpc"
	"github.com/renaynay/astria-hackathon/messenger"
	"google.golang.org/grpc"
)

func main() {
	println("hello, world!")
	m := messenger.NewMessenger()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	execution_server := messenger.NewExecutionServiceServerV1Alpha2(m)
	server := grpc.NewServer()
	astriaGrpc.RegisterExecutionServiceServer(server, execution_server)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
