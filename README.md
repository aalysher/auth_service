# Auth Service

## Description

Auth Service is an authentication microservice implemented in Go, providing user authentication using JWT tokens and password hashing with bcrypt. The microservice utilizes gRPC for communication with other services and manages users through an SQL database running in a Docker container.

## Functionality

- User registration and authentication
- JWT token generation and validation
- User profile retrieval

## Project Structure
```
auth-service/
├── cmd/
│ └── server/
│ └── main.go # Main entry point for the server
├── config/
│ └── config.yml # Configuration file
├── internal/
│ ├── auth/
│ │ └── jwt_manager.go # JWT token management
│ ├── handler/
│ │ └── auth.go # gRPC request handlers
│ └── server/
│ └── server.go # Server setup and initialization
└── proto/
├── auth.proto # gRPC service and message definitions
└── auth.pb.go # Generated gRPC files
```

## Installation

1. **Clone the repository:**

    ```bash
    git clone https://github.com/yourusername/auth-service.git
    cd auth-service
    ```

2. **Install dependencies:**

    ```bash
    go mod tidy
    ```

3. **Configure the settings:**

    Copy the configuration file and adjust it as needed:

    ```bash
    cp config/config.yml.example config/config.yml
    ```

4. **Run Docker containers for the database:**

    Create a `docker-compose.yml` file for your database, for example:

    ```yaml
    version: '3'
    services:
      db:
        image: postgres:latest
        environment:
          POSTGRES_USER: user
          POSTGRES_PASSWORD: password
          POSTGRES_DB: authdb
        ports:
          - "5432:5432"
    ```

    Start the containers:

    ```bash
    docker-compose up -d
    ```

5. **Generate gRPC files:**

    Ensure `protoc` is installed, and run:

    ```bash
    protoc --go_out=. --go-grpc_out=. proto/auth.proto
    ```

6. **Run the server:**

    ```bash
    go run cmd/server/main.go
    ```

## Usage

### Authentication and Profile Retrieval

Use a gRPC client to interact with the service. For example, using `grpcurl`:

- **Authentication:**

    ```bash
    grpcurl -d '{"username": "testuser", "password": "testpass"}' -H "Authorization: Bearer <TOKEN>" localhost:50051 auth.AuthService/Login
    ```

- **Get User Profile:**

    ```bash
    grpcurl -d '{"user_id": "user-id"}' -H "Authorization: Bearer <TOKEN>" localhost:50051 auth.AuthService/GetUserProfile
    ```

## Testing

You can use Go testing libraries to unit test your code. Create tests in the `internal/handler/` directory and run them:

    ```bash
    go test ./internal/handler
    ```



## Development

1. **Run the server in development mode:**
    ```bash
        go run cmd/server/main.go
    ```

2. **To add new features:**
    - Modify files in proto/, then regenerate gRPC files.
    - Update internal/handler/ and other components to handle new features.
3. **Commit and push changes:**
    ```bash
    git add .
    git commit -m "Added new feature"
    git push origin main
    ```

## License
This project is licensed under the MIT License. See the LICENSE file for details.
