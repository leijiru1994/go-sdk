package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/leijjiru1994/go-sdk/infrastructure/tracer"

	"github.com/DeanThompson/ginpprof"
	"github.com/facebookgo/grace/gracenet"
	"github.com/gin-gonic/gin"
	zipkinHttp "github.com/openzipkin/zipkin-go/middleware/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

func NewDefaultServer(ctx context.Context, conf *Config) (server *Server, err error) {
	err = tracer.InitTracer(conf.Tracer)
	if err != nil {
		return
	}

	engine := gin.New()
	engine.Use(conf.HandlerFunc...)
	engine.GET("metrics", gin.WrapH(promhttp.Handler()))
	ginpprof.Wrap(engine)

	if conf.ReadTimeout == 0 {
		conf.ReadTimeout = time.Second
	}

	if conf.WriteTimeout == 0 {
		conf.WriteTimeout = time.Second*10
	}

	return &Server{
		HttpAddr: conf.GetHttpAddr(),
		PidFile:  conf.PidFile,
		Engine:   engine,
		S:        nil,
		net:      &gracenet.Net{},
		CanKill:  conf.CanKill,

		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		Cors:         conf.Cors,
	}, nil
}

func (s *Server) Run(ctx context.Context) (err error) {
	s.S = &http.Server{
		Addr:           s.HttpAddr,
		Handler:        cors.New(s.Cors).Handler(zipkinHttp.NewServerMiddleware(tracer.GlobalTracer)(s.Engine)),
		ReadTimeout:    s.ReadTimeout,
		WriteTimeout:   s.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	var httpListener net.Listener
	httpListener, err = s.net.Listen("tcp", s.HttpAddr)
	if err != nil {
		return
	}

	errC := make(chan error, 1)
	go func() {
		errC <- s.S.Serve(httpListener)
	}()


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
	err = s.S.Shutdown(context.Background())
	return
}
