# API Audit Logging

## Overview

The API audit logging system automatically logs all API requests (GET, POST, PUT, DELETE, PATCH, etc.) to the database for comprehensive tracking and auditing purposes.

## Architecture

### Database Collection

All audit logs are stored in the `audit_logs` collection in MongoDB.

### Data Model

Each audit log entry contains:

- `id`: Unique identifier (auto-generated)
- `method`: HTTP method (GET, POST, PUT, DELETE, PATCH, etc.)
- `path`: Request path/endpoint (e.g., `/api/users/all`)
- `userId`: User ID if authenticated (optional)
- `userEmail`: User email if authenticated (optional)
- `ipAddress`: Client IP address
- `userAgent`: User agent string
- `statusCode`: HTTP response status code
- `requestTime`: Timestamp when request was made
- `duration`: Request duration in milliseconds
- `requestBody`: Request body (sanitized - sensitive fields redacted)
- `queryParams`: Query parameters
- `error`: Error message if any
- `responseSize`: Response body size in bytes
- `created`: Timestamp from Model
- `updated`: Optional update timestamp from Model

## Features

### Automatic Logging

- **All API requests** are automatically logged (GET, POST, PUT, DELETE, PATCH, etc.)
- Logging happens **asynchronously** to avoid blocking API responses
- Sensitive fields (passwords, tokens, secrets) are automatically **redacted**

### Excluded Endpoints

The following endpoints are excluded from logging:

- `/api/health` - Health check endpoint
- `/api/metrics/*` - Metrics endpoints
- `/ws` - WebSocket connections
- `/uploads/*` - Static file serving

### Security

- **Sensitive data sanitization**: Fields containing "password", "token", "secret", "accessToken", "refreshToken", or "authorization" are automatically redacted
- **Recursive sanitization**: Nested objects are also sanitized
- **No blocking**: Logging failures don't affect API responses

## Usage

### Middleware Integration

The audit logging middleware is automatically integrated into the main router:

```go
router.Use(middleware.AuditLogMiddleware())
```

### Repository Operations

The `AuditLogRepo` provides several query methods:

```go
repo := repos.NewAuditLogRepo()

// Get logs by user ID
logs, err := repo.GetByUserID(ctx, userID, &limit)

// Get logs by path/endpoint
logs, err := repo.GetByPath(ctx, "/api/users/all", &limit)

// Get logs by HTTP method
logs, err := repo.GetByMethod(ctx, "POST", &limit)

// Get logs by status code
logs, err := repo.GetByStatusCode(ctx, 404, &limit)

// Get recent logs with filters
filters := bson.M{"method": "POST", "statusCode": 200}
logs, err := repo.GetRecent(ctx, 100, filters)

// Get logs by date range
logs, err := repo.GetByDateRange(ctx, startDate, endDate, &limit)
```

## Database Indexes

The following indexes are created for optimal query performance:

- `id` (unique) - Primary key
- `userId` - For querying by user
- `path` - For querying by endpoint
- `method` - For querying by HTTP method
- `requestTime` (descending) - For sorting and date range queries
- `statusCode` - For querying by status code

To create indexes, run:

```bash
go run scripts/create_mongodb_indexes.go
```

## Query Examples

### Get all POST requests from a specific user

```go
repo := repos.NewAuditLogRepo()
filters := bson.M{
    "userId": userID,
    "method": "POST",
}
logs, err := repo.GetRecent(ctx, 100, filters)
```

### Get all failed requests (4xx, 5xx)

```go
repo := repos.NewAuditLogRepo()
filters := bson.M{
    "statusCode": bson.M{"$gte": 400},
}
logs, err := repo.GetRecent(ctx, 100, filters)
```

### Get all requests to a specific endpoint in the last 24 hours

```go
repo := repos.NewAuditLogRepo()
startDate := time.Now().Add(-24 * time.Hour)
endDate := time.Now()
logs, err := repo.GetByDateRange(ctx, startDate, endDate, nil)
// Then filter by path in application code or use aggregation
```

## Performance Considerations

1. **Asynchronous Logging**: Logs are saved in a goroutine to avoid blocking responses
2. **Indexes**: Essential indexes are created for common query patterns
3. **Selective Logging**: Health checks and metrics are excluded to reduce noise
4. **Body Size Limiting**: Large request/response bodies may impact performance - consider adding size limits if needed

## Future Enhancements

Potential improvements:

1. **Log Rotation**: Implement automatic cleanup of old logs (e.g., keep logs for 90 days)
2. **Aggregation Queries**: Add pre-aggregated statistics (requests per hour, error rates, etc.)
3. **Alerting**: Integrate with alerting system for suspicious patterns
4. **Export**: Add functionality to export logs for compliance/analysis
5. **Search**: Add full-text search capabilities for request bodies
6. **Rate Limiting**: Add rate limiting based on audit log patterns

## Notes

- Audit logs are separate from `activity_logs` which track entity-specific changes (tasks, projects, users)
- Audit logs track all API requests regardless of success/failure
- The system is designed to be non-intrusive - logging failures don't affect application functionality

