package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/labstack/echo/v4"

	// "github.com/labstack/echo/v4/middleware"
	// "github.com/labstack/echo/v4/middleware"
	util "go-dynamic-grpc/utils" // Replace with your actual module path

	"google.golang.org/grpc"
)

// this is for the api response only
type ApiResponse struct {
	Message string `json:"message"`
}

func handleHello(c echo.Context) error {
	conn, err := grpc.Dial(
		"132.186.123.91:50051",
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
	var requestBody map[string]interface{}
	if err := c.Bind(&requestBody); err != nil {
		fmt.Println("Failed to bind request body:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	fmt.Println("Request body:", requestBody)

	// Check if the required fields are present and of correct type
	methodRaw, methodOk := requestBody["method"].(string)
	inputRaw, inputOk := requestBody["input"]
	if !methodOk || !inputOk {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing or invalid 'method' or 'input' field in request body"})
	}
	methodName := methodRaw

	// Convert inputRaw back to JSON bytes (inputRaw is usually a map[string]interface{})
	inputJSON, err := json.Marshal(inputRaw)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to marshal 'input' field to JSON"})
	}

	// Get the proto map (if needed for dynamic method discovery)
	protoMap := getProtoMap()

	methodInfo, found := protoMap[methodName]
	if !found {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unknown method: " + methodName})
	}

	res, err := invokeDynamic(conn, methodInfo.MethodDescriptor, inputJSON)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("gRPC error: %v", err)})
	}

	// response in json format
	if res == nil {
		return c.JSON(http.StatusOK, map[string]string{"message": "No response from gRPC server"})
	}	
	// If the response is not nil, we can assume it's a valid response
	jsonResponse := make(map[string]interface{})
	if err := json.Unmarshal(res, &jsonResponse); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal gRPC response"})
	}
	// Convert the response to ApiResponse
	jsonResponseBytes, err := json.Marshal(jsonResponse)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal gRPC response to string"})
	}
	apiResponse := ApiResponse{
		Message: string(jsonResponseBytes),
	}

	return c.JSON(http.StatusOK, apiResponse)
}

func main() {
	e := echo.New()
	// e.Use(middleware.BodyLimit("200M"))

	e.POST("/hello", handleHello)

	e.Logger.Fatal(e.Start(":3005"))
}

func getProtoMap() map[string]*util.GrpcMethodInfo {
	protoMap, err := util.ParseProtoAndBuildMap(
		[]string{"proto/hello.proto"},
		[]string{""},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse proto files: %v", err))
	}

	fmt.Println("Proto map built successfully with", len(protoMap), protoMap)
	return protoMap
}

func invokeDynamic(
	conn *grpc.ClientConn,
	methodDesc *desc.MethodDescriptor,
	inputJSON []byte,
) ([]byte, error) {
	// Create a dynamic request from descriptor
	req := dynamic.NewMessage(methodDesc.GetInputType())

	// Unmarshal the JSON input into the request message
	if err := req.UnmarshalJSON(inputJSON); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}

	// Create a dynamic stub using grpc.ClientConn
	methodName := fmt.Sprintf("/%s/%s",
		methodDesc.GetService().GetFullyQualifiedName(),
		methodDesc.GetName(),
	)

	// Prepare the response message dynamically
	res := dynamic.NewMessage(methodDesc.GetOutputType())

	// Do the gRPC call
	err := conn.Invoke(context.Background(), methodName, req, res)
	if err != nil {
		return nil, fmt.Errorf("gRPC error: %w", err)
	}

	// Convert the result to JSON
	return res.MarshalJSON()
}