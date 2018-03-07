package grpc

import (
	_ "golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"

)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func NewSlave()  {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
}