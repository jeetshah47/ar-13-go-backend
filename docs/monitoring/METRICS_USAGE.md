# Service Metrics and Resource Usage Monitoring

This document explains how to use the built-in metrics system to identify which microservices are consuming the most resources on your EC2 instance.

## Overview

The application now includes a comprehensive metrics tracking system that monitors:
- **Request Count**: Number of requests per endpoint/service
- **Response Time**: Duration of requests (min, max, average)
- **Network Usage**: Bytes in/out per endpoint/service
- **CPU Impact**: Indirectly measured through response time

## Metrics Endpoints

### 1. Get All Metrics by Endpoint
```
GET /api/metrics/all
```
Returns detailed metrics for all endpoints, sorted by request count.

**Response Example:**
```json
{
  "endpoints": [
    {
      "endpoint": "/api/tasks/all/:id",
      "service": "tasks",
      "request_count": 1250,
      "total_duration_ms": 125000,
      "avg_duration_ms": 100.0,
      "min_duration_ms": 50,
      "max_duration_ms": 500,
      "total_bytes_in": 5000000,
      "total_bytes_out": 25000000,
      "avg_bytes_in": 4000.0,
      "avg_bytes_out": 20000.0,
      "last_request_time": "2024-01-15T12:00:00Z"
    }
  ],
  "total_endpoints": 45
}
```

### 2. Get Metrics Grouped by Service
```
GET /api/metrics/by-service
```
Returns metrics aggregated by service (e.g., tasks, projects, users, etc.).

**Response Example:**
```json
{
  "services": [
    {
      "service": "tasks",
      "request_count": 5000,
      "total_duration_ms": 500000,
      "avg_duration_ms": 100.0,
      "total_bytes_in": 20000000,
      "total_bytes_out": 100000000,
      "avg_bytes_in": 4000.0,
      "avg_bytes_out": 20000.0,
      "endpoints": [...]
    }
  ],
  "total_services": 15
}
```

### 3. Get Top Services by Resource Usage
```
GET /api/metrics/top?sort_by=requests&limit=10
```
Returns the top N services sorted by various criteria.

**Query Parameters:**
- `sort_by`: Sort criteria (default: `requests`)
  - `requests`: Total request count
  - `duration`: Total response time
  - `bytes_in`: Total bytes received
  - `bytes_out`: Total bytes sent
  - `total_bytes`: Total network usage (in + out)
- `limit`: Number of results to return (default: 10)

**Response Example:**
```json
{
  "top_services": [
    {
      "service": "tasks",
      "request_count": 5000,
      "total_duration_ms": 500000,
      "total_bytes_in": 20000000,
      "total_bytes_out": 100000000,
      "endpoints": [...]
    }
  ],
  "summary": {
    "total_services": 15,
    "total_requests": 25000,
    "total_duration_ms": 2500000,
    "total_bytes_in": 100000000,
    "total_bytes_out": 500000000,
    "total_bytes": 600000000
  },
  "sort_by": "requests",
  "limit": 10
}
```

### 4. Reset Metrics
```
POST /api/metrics/reset
```
Clears all collected metrics. Useful for starting fresh or periodic resets.

**Note:** Consider adding authentication to this endpoint in production.

## Correlating with EC2 Metrics

### Understanding EC2 Metrics

From your CloudWatch dashboard, you're seeing:
1. **CPU Utilization (%)**: Overall CPU usage
2. **Network In (bytes)**: Data received by the instance
3. **Network Out (bytes)**: Data sent from the instance
4. **Network Packets In/Out**: Number of network packets
5. **CPU Credit Usage/Balance**: For burstable instances

### How to Correlate

1. **Network Usage Correlation:**
   - Compare `total_bytes_in` from `/api/metrics/top?sort_by=bytes_in` with EC2 "Network In"
   - Compare `total_bytes_out` from `/api/metrics/top?sort_by=bytes_out` with EC2 "Network Out"
   - Services with high `total_bytes_in` or `total_bytes_out` are contributing more to network traffic

2. **CPU Usage Correlation:**
   - Services with high `avg_duration_ms` or `total_duration_ms` are likely using more CPU
   - High request count with high average duration = CPU intensive service
   - Use `/api/metrics/top?sort_by=duration` to find CPU-intensive services

3. **Request Volume Correlation:**
   - High `request_count` services generate more network packets
   - Compare with EC2 "Network Packets In/Out" metrics
   - Use `/api/metrics/top?sort_by=requests` to find high-traffic services

## Example Analysis Workflow

1. **Check Top Services by Network Usage:**
   ```bash
   curl http://your-server:3000/api/metrics/top?sort_by=total_bytes&limit=5
   ```

2. **Check Top Services by Request Count:**
   ```bash
   curl http://your-server:3000/api/metrics/top?sort_by=requests&limit=5
   ```

3. **Check Top Services by CPU Time:**
   ```bash
   curl http://your-server:3000/api/metrics/top?sort_by=duration&limit=5
   ```

4. **Compare with EC2 Metrics:**
   - If EC2 shows high "Network In" → Check which services have high `total_bytes_in`
   - If EC2 shows high "Network Out" → Check which services have high `total_bytes_out`
   - If EC2 shows high "CPU Utilization" → Check which services have high `total_duration_ms` or `avg_duration_ms`

## Service Identification

Based on your codebase, here are the main services:

- **tasks**: Task management endpoints (likely high usage)
- **project**: Project management endpoints
- **users**: User management endpoints
- **auth**: Authentication endpoints (login, register)
- **dashboard**: Dashboard statistics
- **calendar**: Calendar events
- **notifications**: Notification system
- **vacation**: Vacation requests
- **employee**: Employee management
- **info-portal**: Info portal pages
- **project-details**: Project details
- **activity-log**: Activity logging
- **google-account**: Google OAuth integration
- **drawing-list**: Drawing list management
- **storage**: File storage operations (likely high network usage)

## Recommendations

1. **High Network Usage Services:**
   - `storage`: File uploads/downloads
   - `tasks`: Task attachments and file operations
   - `info-portal`: Page attachments

2. **High CPU Usage Services:**
   - `dashboard`: Statistics calculations
   - `tasks`: Complex task operations
   - `project`: Project statistics

3. **High Request Volume Services:**
   - `auth`: Login/register requests
   - `notifications`: Real-time notification polling
   - `tasks`: Frequent task updates

## Monitoring Best Practices

1. **Regular Monitoring:**
   - Check metrics daily or weekly
   - Reset metrics periodically to see current trends
   - Compare metrics with EC2 CloudWatch data

2. **Performance Optimization:**
   - Focus optimization efforts on services with:
     - High `avg_duration_ms` (slow endpoints)
     - High `total_bytes_out` (large responses)
     - High `request_count` with high `avg_duration_ms` (frequent slow requests)

3. **Scaling Decisions:**
   - Use metrics to identify which services need more resources
   - Consider horizontal scaling for high-traffic services
   - Optimize or cache responses for high `avg_duration_ms` endpoints

## Security Note

The metrics endpoints are currently public. In production, consider:
- Adding authentication middleware
- Restricting access to admin users only
- Using rate limiting on metrics endpoints

