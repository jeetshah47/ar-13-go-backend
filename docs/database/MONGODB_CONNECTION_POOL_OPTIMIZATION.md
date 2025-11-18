# MongoDB Connection Pool Optimization

## Problem

MongoDB connections were increasing above 20, sometimes reaching high numbers. This was causing:
- Unnecessary resource consumption
- Potential connection limit issues on MongoDB Atlas free tier
- Increased latency due to connection overhead

## Root Causes

### 1. Excessive MaxPoolSize
- **Previous Setting**: `MaxPoolSize = 100`
- **Issue**: Allowed up to 100 concurrent connections, which is excessive for most applications
- **Impact**: Connections could grow unnecessarily during traffic spikes

### 2. High MinPoolSize
- **Previous Setting**: `MinPoolSize = 10`
- **Issue**: Kept 10 connections always open, even during low traffic
- **Impact**: Wasted resources during idle periods

### 3. Short Idle Timeout
- **Previous Setting**: `MaxConnIdleTime = 30 seconds`
- **Issue**: Connections closed too quickly, causing frequent reconnections
- **Impact**: Increased connection churn and overhead

### 4. Missing Connection Limits
- **Issue**: No limits on concurrent connection attempts
- **Impact**: Could create many connections simultaneously during traffic spikes

### 5. Missing Timeout Settings
- **Issue**: No explicit timeouts for connection establishment and operations
- **Impact**: Connections could hang indefinitely

## Solution

### Optimized Connection Pool Settings

```go
SetMaxPoolSize(20)                    // Maximum connections in pool (reduced from 100)
SetMinPoolSize(2)                     // Minimum connections to maintain (reduced from 10)
SetMaxConnecting(5)                   // Max concurrent connection attempts
SetMaxConnIdleTime(5 * time.Minute)  // Close idle connections after 5 minutes
SetConnectTimeout(10 * time.Second)  // Timeout for establishing connections
SetSocketTimeout(30 * time.Second)   // Timeout for socket operations
SetServerSelectionTimeout(5 * time.Second) // Timeout for server selection
```

### Changes Summary

| Setting | Before | After | Reason |
|---------|--------|-------|--------|
| MaxPoolSize | 100 | 20 | Prevents excessive connections |
| MinPoolSize | 10 | 2 | Reduces idle connection overhead |
| MaxConnecting | Not set | 5 | Limits concurrent connection attempts |
| MaxConnIdleTime | 30s | 5 minutes | Reduces connection churn |
| ConnectTimeout | Not set | 10s | Prevents hanging connections |
| SocketTimeout | Not set | 30s | Prevents hanging operations |
| ServerSelectionTimeout | Not set | 5s | Faster failure detection |

## Expected Behavior

### Connection Count
- **Normal Operation**: 2-5 connections (min pool + active operations)
- **Peak Traffic**: Up to 20 connections (max pool size)
- **Idle Periods**: 2 connections (min pool size)

### Connection Lifecycle
1. **Startup**: Creates 2 connections (min pool)
2. **Traffic Increase**: Grows to handle load (up to 20 max)
3. **Traffic Decrease**: Idle connections closed after 5 minutes
4. **Shutdown**: All connections properly closed

## Monitoring

A new function `GetConnectionPoolStats()` has been added to monitor connection pool:

```go
stats, err := mongodb.GetConnectionPoolStats()
// Returns:
// {
//   "sessionsInProgress": <number>,
//   "clientInitialized": <bool>
// }
```

## Best Practices

### 1. Context Management
Always use contexts with timeouts for MongoDB operations:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
// Use ctx in MongoDB operations
```

### 2. Proper Cleanup
Ensure connections are closed on application shutdown:
```go
defer mongodb.Close()
```

### 3. Connection Reuse
The MongoDB client is singleton - all repositories share the same connection pool.

### 4. Monitoring
Regularly check connection pool stats to ensure healthy operation.

## Troubleshooting

### If connections still exceed 20:

1. **Check for connection leaks**:
   - Ensure all contexts are properly cancelled
   - Verify no long-running operations without timeouts
   - Check for goroutines that don't complete

2. **Review concurrent operations**:
   - Limit concurrent API requests if needed
   - Use connection pooling at application level

3. **Monitor MongoDB Atlas**:
   - Check connection count in Atlas dashboard
   - Verify no other applications using same cluster
   - Check for connection spikes during specific operations

4. **Adjust settings if needed**:
   - For high-traffic applications, increase MaxPoolSize (but keep it reasonable)
   - For low-traffic applications, reduce MinPoolSize to 1

## Migration Notes

- **No code changes required**: The optimization is transparent to application code
- **Immediate effect**: New connections will use optimized settings
- **Existing connections**: Will gradually adopt new settings as they reconnect
- **Restart recommended**: For immediate effect, restart the application

## References

- [MongoDB Go Driver Connection Pooling](https://www.mongodb.com/docs/drivers/go/current/fundamentals/connection-pooling/)
- [MongoDB Atlas Connection Limits](https://www.mongodb.com/docs/atlas/reference/connection-limits/)

