package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/leijjiru1994/go-sdk/grpc/interceptor/server/access_log"
	"github.com/leijjiru1994/go-sdk/grpc/interceptor/server/metric"
	"github.com/leijjiru1994/go-sdk/grpc/interceptor/server/trace"
	"github.com/leijjiru1994/go-sdk/infrastructure/log"
	"github.com/leijjiru1994/go-sdk/infrastructure/tracer"

	"github.com/facebookgo/grace/gracenet"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func NewDefaultServer(ctx context.Context, conf *Config) (server *Server, err error) {
	var logger zerolog.Logger
	logger, err = log.InitLogger(ctx, &log.Config{
		BizName: "access_log",
		File:    conf.AccessLog,
		Level:   conf.LogLevel,
	})
	if err != nil {
		return
	}

	opts := grpc.ChainUnaryInterceptor(
		metric.UnaryServerInterceptor(),
		trace.UnaryServerInterceptor(tracer.GlobalTracer),
		access_log.UnaryServerInterceptor(access_log.WithLogger(logger)),
		recovery.UnaryServerInterceptor())
	return NewServer(conf, opts), nil
}

func NewServer(conf *Config, opt ...grpc.ServerOption) *Server {
	s1 := grpc.NewServer(opt...)
	smux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
			UseEnumNumbers: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}))

	return &Server{
		HttpAddr: conf.GetHttpAddr(),
		GrpcAddr: conf.GetGRPCAddr(),
		PidFile:  conf.PidFile,
		S1:       s1,
		S2:       nil,
		net:      &gracenet.Net{},
		SMux:     smux,
		CanKill:  conf.CanKill,
	}
}

func (s *Server) Run(
	ctx context.Context,
	desc *grpc.ServiceDesc,
	impl interface{},
	srvHandlers []ServiceHandler) (err error) {
	var grpcListener, httpListener net.Listener
	grpcListener, err = s.net.Listen("tcp", s.GrpcAddr)
	if err != nil {
		return
	}

	httpListener, err = s.net.Listen("tcp", s.HttpAddr)
	if err != nil {
		return
	}

	s.S1.RegisterService(desc, impl)
	grpc_prometheus.Register(s.S1)

	errC := make(chan error, 1)
	go func() {
		errC <- s.S1.Serve(grpcListener)
	}()
	go func() {
		errC <- s.S2.Serve(httpListener)
	}()

	// grpc http gateway
	conn, _ := grpc.DialContext(
		ctx,
		s.GrpcAddr,
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)

	if srvHandlers != nil && len(srvHandlers) > 0 {
		for _, registerFunc := range srvHandlers {
			err = registerFunc(ctx, s.SMux, conn)
			if err != nil {
				return
			}
		}
	}

	if s.CanKill {
		killPPid()
	}
	_ = overwritePid(s.PidFile)
	quit := s.handleSignal(errC)
	select {
	case tmpErr := <-errC:
		return tmpErr
	case <-quit:
		return
	}
}

func (s *Server) handleSignal(errs chan error) <-chan struct{} {
	quit := make(chan struct{})
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

		for sig := range ch {
			switch sig {
			// stop service
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				signal.Stop(ch)
				_ = s.Stop()
				close(quit)

				return

			// restart service
			case syscall.SIGUSR1, syscall.SIGUSR2:
				if _, tmpErr := s.net.StartProcess(); tmpErr != nil {
					errs <- tmpErr
				}
			}
		}
	}()

	return quit
}

func killPPid() {
	ppid := os.Getppid()
	if ppid == 1 {
		return
	}

	_ = syscall.Kill(ppid, syscall.SIGTERM)
}

func overwritePid(fileName string) (err error) {
	if fileName == "" {
		return
	}

	var f *os.File
	f, err = os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return
	}

	n, _ := f.Seek(0, os.SEEK_END)
	_, err = f.WriteAt([]byte(strconv.Itoa(os.Getpid())), n)
	_ = f.Close()

	return
}

func (s *Server) Stop() (err error) {
	s.S1.GracefulStop()
	err = s.S2.Shutdown(context.Background())
	return
}
