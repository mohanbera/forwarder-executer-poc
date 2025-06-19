package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"

	pb "go-forwarder/pb" // Replace with your actual module path
)

// Greeter server implementation
type greeterServer struct {
	pb.UnimplementedGreeterServer
}

// StreamHello sends multiple greetings over time
func (s *greeterServer) StreamHello(req *pb.HelloRequest, stream pb.Greeter_StreamHelloServer) error {
	name := req.GetName()
	for i := 1; i <= 5; i++ {
		msg := fmt.Sprintf("Hello %s #%d", name, i)
		if err := stream.Send(&pb.HelloReply{Message: msg}); err != nil {
			return err
		}
		time.Sleep(1 * time.Second) // Simulate delay
	}
	return nil
}

func main() {
	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterGreeterServer(s, &greeterServer{})
		log.Println("gRPC server listening on :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start Echo server
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from Echo HTTP Server!")
	})
	log.Println("Echo HTTP server listening on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
