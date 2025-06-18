const grpc = require("@grpc/grpc-js");
const protoLoader = require("@grpc/proto-loader");
const path = require("path");

// Load proto file
const packageDefinition = protoLoader.loadSync(
  path.join(__dirname, "proto", "hello.proto"),
  {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true,
  }
);

const helloProto = grpc.loadPackageDefinition(packageDefinition).hello;

// Implement service
const HelloService = {
  SayHello: (call, callback) => {
    const name = call.request.name;
    callback(null, { message: `Hello, ${name}!` });
  },
  SayGoodbye: (call, callback) => {
    const name = call.request.name;
    callback(null, { message: `Goodbye, ${name}!` });
  },
};

// Start server
function main() {
  const server = new grpc.Server({
    "grpc.max_receive_message_length": 200 * 1024 * 1024,
    "grpc.max_send_message_length": 200 * 1024 * 1024,
  });
  server.addService(helloProto.HelloService.service, HelloService);
  server.bindAsync(
    "0.0.0.0:50051",
    grpc.ServerCredentials.createInsecure(),
    () => {
      server.start();
    }
  );
}

main();
