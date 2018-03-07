package grpc

import (
	"golang.org/x/net/context"
	gogrpc "google.golang.org/grpc"
    "github.com/golang/protobuf/ptypes/empty"
	"goest-worker/common"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

type Master struct{
	backend common.PoolBackendInterface
	server *gogrpc.Server
}

type MasterService struct {}


func (master *MasterService) AddSlave(context.Context, *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (master *MasterService) JobStream(stream Master_JobStreamServer) error {
	go func() {
		for {
			result, err := stream.Recv()
			if err != nil{
				return
			}
			for _, arg := range result.Args {
				arg.Value.GetValue()
			}
		}
	}()
	go func() {}()
	return nil
}

func NewMasterServer(addr string, backend common.PoolBackendInterface) (*Master, error) {
	s := gogrpc.NewServer()
	RegisterMasterServer(s, &MasterService{})
	return &Master{
		server: s,
		backend: backend,
	}, nil
}