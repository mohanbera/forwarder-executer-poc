package main

import (
	"context"
	"encoding/json"
	"fmt"
	hello "go-dynamic-grpc/pb/hello" // Replace with your actual module path
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"
	// "github.com/labstack/echo/v4/middleware"
	// "github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"
)

// this is for the api response only
type ApiResponse struct {
	Message string `json:"message"`
}

func handleHello(c echo.Context) error {
	conn, err := grpc.Dial(
		"node-server:50051",
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(200*1024*1024),
			grpc.MaxCallRecvMsgSize(200*1024*1024),
		),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to gRPC server"})
	}
	defer conn.Close()

	// Map service and method to client and message types
	clientMap := map[string]interface{}{
		"hello": hello.NewHelloServiceClient(conn),
	}
	requestTypeMap := map[string]reflect.Type{
		"hello.SayHello": reflect.TypeOf((*hello.HelloRequest)(nil)).Elem(),
	}
	// responseTypeMap := map[string]reflect.Type{
	// 	"hello.SayHello": reflect.TypeOf((*hello.HelloResponse)(nil)).Elem(),
	// }

	// Example data (usually from the request)
	serviceName := "hello"
	methodName := "SayHello"
	fullMethodKey := serviceName + "." + methodName
	message := map[string]interface{}{"name": "Hello, gRPC!"} // from JSON body, ideally

	// Create the request struct dynamically
	reqType, ok := requestTypeMap[fullMethodKey]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unknown method"})
	}
	reqPtr := reflect.New(reqType) // *HelloRequest

	// Populate request from generic map (e.g. JSON)
	reqJSON, _ := json.Marshal(message)
	if err := json.Unmarshal(reqJSON, reqPtr.Interface()); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Dynamically call the method
	client := clientMap[serviceName]
	res, err := callGrpcMethodWithArg(client, methodName, context.Background(), reqPtr.Interface())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Handle the response (optional: use reflection to inspect it if needed)
	respJSON, _ := json.Marshal(res)
	return c.JSON(http.StatusOK, json.RawMessage(respJSON))
}

func main() {
	e := echo.New()
	// e.Use(middleware.BodyLimit("200M"))

	e.POST("/hello", handleHello)

	e.Logger.Fatal(e.Start(":3005"))
}

func callGrpcMethodWithArg(obj interface{}, methodName string, ctx context.Context, req interface{}) (interface{}, error) {
	v := reflect.ValueOf(obj)
	method := v.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(req),
	}

	results := method.Call(args)
	if len(results) != 2 {
		return nil, fmt.Errorf("unexpected number of return values")
	}

	if errInterface := results[1].Interface(); errInterface != nil {
		return nil, errInterface.(error)
	}
	return results[0].Interface(), nil
}