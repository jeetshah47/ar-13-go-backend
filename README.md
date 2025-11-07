# Go Migration - AR-13 Backend

This is the Go (Golang) migration of the AR-13 Node.js backend.

## Project Structure

```
migration/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── constants/           # Constants (paths, HTTP status codes)
│   ├── handlers/            # HTTP request handlers
│   ├── middleware/          # HTTP middleware (auth, CORS, etc.)
│   ├── models/              # Data models
│   ├── repos/               # Data access layer (Firestore)
│   └── services/            # Business logic
├── pkg/
│   ├── auth/                # Authentication utilities
│   ├── email/               # Email service
│   ├── fileupload/          # File upload handling
│   ├── firebase/            # Firebase integration
│   └── websocket/           # WebSocket service
├── upload/                  # Uploaded files directory
│   ├── tasks/
│   └── info-portal/
└── go.mod                    # Go module definition
```

## Setup

1. **Install Go** (version 1.21 or higher)

2. **Install Air (for live reload during development):**
   ```bash
   go install github.com/air-verse/air@latest
   ```
   Make sure `$HOME/go/bin` is in your PATH, or use the full path to the air binary.

3. **Install dependencies:**
   ```bash
   go mod download
   ```

4. **Set up environment variables:**
   Create a `.env` file in the root directory (copy from `.env.example`):
   ```
   PORT=3000
   NODE_ENV=development
   
   # AWS DynamoDB
   AWS_REGION=us-east-1
   
   # JWT Configuration
   JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
   JWT_EXPIRATION_HOURS=24
   REFRESH_EXPIRATION_DAYS=30
   
   # Email Configuration
   EMAIL_HOST=smtp.example.com
   EMAIL_PORT=587
   EMAIL_USER=your-email@example.com
   EMAIL_PASSWORD=your-email-password
   EMAIL_FROM=noreply@example.com
   EMAIL_FROM_NAME=AR-13
   FRONTEND_URL=http://localhost:3000
   GOOGLE_CLIENT_ID=your-google-client-id
   GOOGLE_CLIENT_SECRET=your-google-client-secret
   
   # AWS Credentials (for local development)
   # For EC2, use IAM role instead
   # AWS_ACCESS_KEY_ID=your-access-key
   # AWS_SECRET_ACCESS_KEY=your-secret-key
   ```
   
   **Important Notes:**
   - **JWT_SECRET**: Use a strong random string (at least 32 characters)
   - **AWS_REGION**: Should match your DynamoDB tables region (default: us-east-1)
   - **AWS Credentials**: For local development, set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY. For EC2 deployment, use IAM roles instead.

   **Important Notes for Google OAuth:**
   - **GOOGLE_CLIENT_ID**: Should be in the format `123456789-abcdefghijklmnop.apps.googleusercontent.com` (found in Google Cloud Console → APIs & Services → Credentials)
   - **GOOGLE_CLIENT_SECRET**: Should start with `GOCSPX-` (this is NOT the Client ID)
   - **Common Mistake**: Do NOT use the Client Secret value for `GOOGLE_CLIENT_ID`. They are different values.
   - Make sure your OAuth redirect URI is configured in Google Cloud Console to match your callback URL (e.g., `http://localhost:3000/api/google-account/auth/callback`)

5. **Run the server:**
   
   **Option 1: Using Air (recommended for development - auto-reload on file changes):**
   ```bash
   air
   ```
   
   **Option 2: Standard Go run:**
   ```bash
   go run cmd/server/main.go
   ```
   
   > **Note:** Air automatically restarts the server when you save any `.go` file. Install Air with: `go install github.com/air-verse/air@latest`

## Build

Build the binary:
```bash
go build -o bin/server cmd/server/main.go
```

Run the binary:
```bash
./bin/server
```

## Development Status

### ✅ Completed
- Project structure
- Configuration management
- Constants (paths, HTTP status codes)
- Basic server setup
- Middleware (CORS, Auth skeleton)
- Firebase integration setup
- WebSocket service structure
- Handler structure

### 🚧 In Progress
- Model conversions
- Service implementations
- Repository implementations
- Route handlers

### 📋 TODO
- Complete all route handlers
- Implement all services
- Implement all repositories
- Add validation
- Add error handling
- Add logging
- Add tests
- Add documentation

## Key Differences from Node.js Version

1. **Type Safety**: Go provides compile-time type checking
2. **Performance**: Better concurrency with goroutines
3. **Memory**: Lower memory footprint (~160 MB vs ~420 MB)
4. **Deployment**: Single binary (no runtime dependencies)
5. **Error Handling**: Explicit error handling (no exceptions)

## Migration Progress

This is an ongoing migration. The structure is in place, and core components are being converted from TypeScript/Node.js to Go.

## API Documentation

- [Employee API Documentation](./docs/EMPLOYEE_API.md) - Employee endpoints for task counts and statistics

## Notes

- Firebase integration uses the official Go SDK
- WebSocket uses gorilla/websocket library
- HTTP framework is Gin (similar to Express.js)
- File uploads handled with Gin's multipart form handling

