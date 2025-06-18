mod hello {
    tonic::include_proto!("hello");
}
use axum_extra::extract::Multipart;
use axum::extract::DefaultBodyLimit;
use axum::{
    routing::post, // Changed from 'get' to 'post'
    Router, Json, http::StatusCode, response::{IntoResponse, Response}
};
use hello::{hello_service_client::HelloServiceClient, HelloRequest};
use serde::Serialize;
use tonic::transport::Channel;

#[derive(Serialize)]
struct ApiResponse {
    message: String,
}

// Custom error type for better error handling
enum AppError {
    GrpcError(tonic::Status),
    MultipartError(String),
    IoError(String),
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, error_message) = match self {
            AppError::GrpcError(status) => (StatusCode::INTERNAL_SERVER_ERROR, format!("gRPC error: {}", status)),
            AppError::MultipartError(msg) => (StatusCode::BAD_REQUEST, msg),
            AppError::IoError(msg) => (StatusCode::INTERNAL_SERVER_ERROR, msg),
        };
        (status, Json(serde_json::json!({ "error": error_message }))).into_response()
    }
}


// The handler now accepts a Multipart payload and returns a Result
async fn say_hello_handler(mut multipart: Multipart) -> Result<Json<ApiResponse>, AppError> {
    let mut file_text: Option<String> = None;

    // The form field name for the file should be "upload"
    while let Some(field) = multipart.next_field().await.map_err(|err| AppError::MultipartError(err.to_string()))? {
        if field.name() == Some("upload") {
            let data = field.bytes().await.map_err(|err| AppError::MultipartError(err.to_string()))?;
            
            // Assuming the uploaded file is UTF-8 text.
            // For binary files, you should change your .proto to use `bytes` instead of `string`.
            let text = String::from_utf8(data.to_vec())
                .map_err(|err| AppError::MultipartError(format!("Invalid UTF-8 sequence: {}", err)))?;
            
            file_text = Some(text);
            break; // Found the file, no need to process other fields
        }
    }

    let file_text = file_text.ok_or_else(|| AppError::MultipartError("Missing 'upload' field in form data".to_string()))?;

    println!("Connecting to gRPC server at http://node-server:50051");
    let channel = Channel::from_static("http://node-server:50051")
        .connect()
        .await
        .map_err(|_| AppError::IoError("Failed to connect to gRPC server".to_string()))?;

    let mut client = HelloServiceClient::new(channel)
        .max_encoding_message_size(200 * 1024 * 1024)
        .max_decoding_message_size(200 * 1024 * 1024);

    println!("Sending request with file content to gRPC server...");
    let request = tonic::Request::new(HelloRequest { name: file_text });

    let response = client.say_hello(request).await.map_err(AppError::GrpcError)?.into_inner();
    
    println!("Received response from server.");

    Ok(Json(ApiResponse {
        message: response.message,
    }))
}


#[tokio::main]
async fn main() {
    let app = Router::new()
        // The route is now a POST request handler
        .route("/hello", post(say_hello_handler))
        .layer(DefaultBodyLimit::max(1024*1024*200));

    let listener = tokio::net::TcpListener::bind("0.0.0.0:3000")
        .await
        .unwrap();
        
    println!("Axum server listening on {}", listener.local_addr().unwrap());
    axum::serve(listener, app).await.unwrap();
}
