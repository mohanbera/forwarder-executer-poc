const express = require('express');
const multer = require('multer');
const fs = require('fs');
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const path = require('path');

// Load gRPC proto
const PROTO_PATH = path.join(__dirname, 'proto/hello.proto');
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true
});
const helloProto = grpc.loadPackageDefinition(packageDefinition).hello;

// Setup gRPC client
const client = new helloProto.HelloService(
  'node-server:50051',
  grpc.credentials.createInsecure(),
  {
    'grpc.max_send_message_length': 200 * 1024 * 1024,
    'grpc.max_receive_message_length': 200 * 1024 * 1024,
  }
);

// Express app
const app = express();
const upload = multer({ limits: { fileSize: 200 * 1024 * 1024 } });

app.post('/hello', upload.single('upload'), (req, res) => {
  if (!req.file) {
    return res.status(400).json({ error: "Missing 'upload' field" });
  }

  const text = req.file.buffer.toString();

  client.SayHello({ name: text }, (err, response) => {
    if (err) {
      return res.status(500).json({ error: `gRPC error: ${err.message}` });
    }
    res.json({ message: response.message });
  });
});

app.listen(3002, () => {
  console.log('Listening on port 3002');
});
