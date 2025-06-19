package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	greet "echo-hello/pb/greet" // Adjust the import path
	hello "echo-hello/pb/hello" // Adjust the import path

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"
)

type ApiResponse struct {
	Message string `json:"message"`
}

func sayHelloHandler(c echo.Context) error {
	// Read file from multipart form
	file, err := c.FormFile("upload")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing 'upload' field"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to open uploaded file"})
	}
	defer src.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read file content"})
	}
	text := buf.String()

	// Connect to gRPC server
	conn, err := grpc.Dial(
		"node-server:50051",
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(200*1024*1024), // Max request size
			grpc.MaxCallRecvMsgSize(200*1024*1024), // Max response size
		),
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to gRPC server"})
	}
	defer conn.Close()

	client := hello.NewHelloServiceClient(conn)

	// Send request
	res, err := client.SayHello(context.Background(), &hello.HelloRequest{Name: text})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("gRPC error: %v", err)})
	}

	// Return response
	return c.JSON(http.StatusOK, ApiResponse{Message: res.GetMessage()})
}

func greetHandler(c echo.Context) error {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Connect failed: %v", err)
	}
	defer conn.Close()

	client := greet.NewGreeterClient(conn)

	stream, err := client.StreamHello(context.Background(), &greet.HelloRequest{Name: "Alice"})
	if err != nil {
		log.Fatalf("StreamHello failed: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			break
		}
		log.Println("Received:", msg.GetMessage())
	}

	return c.JSON(http.StatusOK, ApiResponse{Message: "Stream completed"})
}

func main() {
	e := echo.New()
	e.Use(middleware.BodyLimit("200M"))

	e.POST("/hello", sayHelloHandler)
	e.POST("/greet-stream", greetHandler)

	e.Logger.Fatal(e.Start(":3001"))
}
