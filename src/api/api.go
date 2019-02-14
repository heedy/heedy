package api

import (
	"context"
	"fmt"

	"github.com/connectordb/connectordb/api/pb"
)

type API struct{}

func (api *API) SayHello(ctx context.Context, in *pb.PingMessage) (*pb.PingMessage, error) {
	fmt.Printf("Got message %s", in.Greeting)
	return &pb.PingMessage{
		Greeting: "hey there!",
	}, nil
}
