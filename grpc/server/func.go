package server

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type ServiceHandler func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error
