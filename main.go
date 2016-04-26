package main

import (
	"io/ioutil"
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
	schema, err := ioutil.ReadFile("helloworld/helloworld.swagger.json")
	if err != nil {
		return err
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	if err := hw.RegisterGreeterHandlerFromEndpoint(ctx, gwmux, "127.0.0.1:"+grpcPort, opts); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/helloworld/greeter/swagger", func(w http.ResponseWriter, r *http.Request) {
		Set(w, AccessControl{
			Origin:         "*",
			AllowedMethods: []string{"GET", "HEAD", "OPTIONS"},
		})
		Set(w, ContentType("application/json"))

		w.WriteHeader(http.StatusOK)
		w.Write(schema)
	})
	mux.Handle("/", gwmux)

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
