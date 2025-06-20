package utils

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
)

type GrpcMethodInfo struct {
	ServiceName      string
	MethodName       string
	MethodDescriptor *desc.MethodDescriptor
	InputType        *dynamic.Message
}

func ParseProtoAndBuildMap(protoFiles []string, importPaths []string) (map[string]*GrpcMethodInfo, error) {
	parser := protoparse.Parser{
		ImportPaths:           importPaths,
		InferImportPaths:      true,
		IncludeSourceCodeInfo: true,
	}

	fds, err := parser.ParseFiles(protoFiles...)
	if err != nil {
		return nil, err
	}

	methodMap := make(map[string]*GrpcMethodInfo)

	for _, file := range fds {
		for _, svc := range file.GetServices() {
			svcName := svc.GetFullyQualifiedName()
			for _, m := range svc.GetMethods() {
				methodName := m.GetName()
				fullName := svcName + "." + methodName
				inputType := dynamic.NewMessage(m.GetInputType())

				methodMap[fullName] = &GrpcMethodInfo{
					ServiceName: svcName,
					MethodDescriptor: m,
					MethodName:  methodName,
					InputType:   inputType,
				}
			}
		}
	}

	return methodMap, nil
}
