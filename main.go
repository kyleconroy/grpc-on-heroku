package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gengo/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	hw "github.com/kyleconroy/grpc-heroku/helloworld"
)

type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *hw.HelloRequest) (*hw.HelloReply, error) {
	return &hw.HelloReply{Message: "Hello " + in.Name}, nil
}

func startGRPC(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	hw.RegisterGreeterServer(s, &server{})
	return s.Serve(lis)
}

func startHTTP(httpPort, grpcPort string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := hw.RegisterGreeterHandlerFromEndpoint(ctx, mux, "127.0.0.1:"+grpcPort, opts)
	if err != nil {
		return err
	}

	http.ListenAndServe(":"+httpPort, mux)
	return nil
}

func main() {
	errors := make(chan error)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50080"
	}

	if grpcPort == httpPort {
		panic("Can't listen on the same port")
	}

	go func() {
		errors <- startGRPC(grpcPort)
	}()

	go func() {
		errors <- startHTTP(httpPort, grpcPort)
	}()

	for err := range errors {
		log.Fatal(err)
		return
	}
}
