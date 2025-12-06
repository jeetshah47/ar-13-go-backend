# Employee API Documentation

This document describes the Employee APIs for retrieving employee information, task counts, and detailed statistics.

## Base URL

All endpoints are prefixed with `/api/employee`

## Authentication

All endpoints require authentication. Include the authentication token in the request headers:

```
Authorization: Bearer <your-token>
```

## Endpoints

### 1. Get Employee List

Retrieves a list of all employees with their task counts.

**Endpoint:** `GET /api/employee/list`

**Permissions Required:** `employees:read`

**Request:**
```http
GET /api/employee/list
Authorization: Bearer <token>
```

**Response:**
```json
{
  "employees": [
    {
      "userId": "user123",
      "name": "John Doe",
      "email": "john.doe@example.com",
      "backlogTasks": 5,
      "tasksInProgress": 3,
      "tasksInReview": 2,
      "pendingTasks": 1,
      "totalTasks": 15,
      "activeTasks": 11
    },
    {
      "userId": "user456",
      "name": "Jane Smith",
      "email": "jane.smith@example.com",
      "backlogTasks": 2,
      "tasksInProgress": 4,
      "tasksInReview": 1,
      "pendingTasks": 0,
      "totalTasks": 12,
      "activeTasks": 7
    }
  ],
  "totalEmployees": 2
}
```

**Response Fields:**
- `employees`: Array of employee objects
  - `userId`: Unique identifier for the employee
  - `name`: Employee's full name
  - `email`: Employee's email address
  - `backlogTasks`: Number of tasks in backlog/todo status
  - `tasksInProgress`: Number of tasks currently in progress
  - `tasksInReview`: Number of tasks in review status
  - `pendingTasks`: Number of tasks in pending status
  - `totalTasks`: Total number of tasks assigned to the employee
  - `activeTasks`: Number of active tasks (not completed or cancelled)
- `totalEmployees`: Total number of employees in the system

**Example:**
```bash
curl -X GET "https://api.example.com/api/employee/list" \
  -H "Authorization: Bearer your-token-here"
```

---

### 2. Get Employee Task Counts

Retrieves task counts for a specific employee.

**Endpoint:** `GET /api/employee/task-counts/:userId`

**Permissions Required:** `employees:read`

**Path Parameters:**
- `userId` (required): The unique identifier of the employee

**Request:**
```http
GET /api/employee/task-counts/user123
Authorization: Bearer <token>
```

**Response:**
```json
{
  "employee": {
    "userId": "user123",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "backlogTasks": 5,
    "tasksInProgress": 3,
    "tasksInReview": 2,
    "pendingTasks": 1,
    "totalTasks": 15,
    "activeTasks": 11
  }
}
```

**Response Fields:**
- `employee`: Employee object with task counts
  - Same fields as described in the Get Employee List endpoint

**Error Responses:**

**404 Not Found:**
```json
{
  "message": "Employee not found"
}
```

**500 Internal Server Error:**
```json
{
  "message": "Error fetching employee task counts",
  "error": "Error details here"
}
```

**Example:**
```bash
curl -X GET "https://api.example.com/api/employee/task-counts/user123" \
  -H "Authorization: Bearer your-token-here"
```

---

### 3. Get Employee Task Statistics and Analysis

Retrieves detailed task statistics and analysis for an employee based on time period (month, quarter, or year) with optional project filtering.

**Endpoint:** `GET /api/employee/stats/:userId`

**Permissions Required:** `employees:read`

**Path Parameters:**
- `userId` (required): The unique identifier of the employee

**Query Parameters:**
- `period` (required): Time period type - `"month"`, `"quarter"`, or `"year"`
- `periodValue` (required): The specific period value
  - For `month`: Format `"YYYY-MM"` (e.g., `"2024-01"` for January 2024)
  - For `quarter`: Format `"YYYY-QN"` (e.g., `"2024-Q1"` for Q1 2024)
  - For `year`: Format `"YYYY"` (e.g., `"2024"` for year 2024)
- `projectId` (optional): Filter statistics by a specific project ID

**Request Examples:**

**Get monthly statistics:**
```http
GET /api/employee/stats/user123?period=month&periodValue=2024-01
Authorization: Bearer <token>
```

**Get quarterly statistics:**
```http
GET /api/employee/stats/user123?period=quarter&periodValue=2024-Q1
Authorization: Bearer <token>
```

**Get yearly statistics:**
```http
GET /api/employee/stats/user123?period=year&periodValue=2024
Authorization: Bearer <token>
```

**Get statistics filtered by project:**
```http
GET /api/employee/stats/user123?period=month&periodValue=2024-01&projectId=proj456
Authorization: Bearer <token>
```

**Response:**
```json
{
  "stats": {
    "userId": "user123",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "period": "month",
    "periodValue": "2024-01",
    "overall": {
      "totalTasks": 25,
      "completedTasks": 15,
      "activeTasks": 10,
      "backlogTasks": 3,
      "tasksInProgress": 4,
      "tasksInReview": 2,
      "pendingTasks": 1,
      "totalTimeSpent": 1200,
      "averageTimePerTask": 48.0,
      "completionRate": 60.0
    },
    "byProject": [
      {
        "projectId": "proj123",
        "projectName": "Project Alpha",
        "stats": {
          "totalTasks": 15,
          "completedTasks": 10,
          "activeTasks": 5,
          "backlogTasks": 2,
          "tasksInProgress": 2,
          "tasksInReview": 1,
          "pendingTasks": 0,
          "totalTimeSpent": 800,
          "averageTimePerTask": 53.33,
          "completionRate": 66.67
        }
      },
      {
        "projectId": "proj456",
        "projectName": "Project Beta",
        "stats": {
          "totalTasks": 10,
          "completedTasks": 5,
          "activeTasks": 5,
          "backlogTasks": 1,
          "tasksInProgress": 2,
          "tasksInReview": 1,
          "pendingTasks": 1,
          "totalTimeSpent": 400,
          "averageTimePerTask": 40.0,
          "completionRate": 50.0
        }
      }
    ],
    "byTime": [
      {
        "period": "2024-01",
        "periodLabel": "January 2024",
        "stats": {
          "totalTasks": 25,
          "completedTasks": 15,
          "activeTasks": 10,
          "backlogTasks": 3,
          "tasksInProgress": 4,
          "tasksInReview": 2,
          "pendingTasks": 1,
          "totalTimeSpent": 1200,
          "averageTimePerTask": 48.0,
          "completionRate": 60.0
        }
      }
    ],
    "analysis": {
      "productivityTrend": "increasing",
      "mostActiveProject": "Project Alpha",
      "mostActiveProjectId": "proj123",
      "averageCompletionTime": 1440.5,
      "peakProductivityMonth": "January 2024",
      "taskDistribution": {
        "completed": 15,
        "inprogress": 4,
        "backlog": 3,
        "inreview": 2,
        "pending": 1
      }
    }
  }
}
```

**Response Fields:**

**Top Level:**
- `userId`: Employee's unique identifier
- `name`: Employee's full name
- `email`: Employee's email address
- `period`: The time period type used (`"month"`, `"quarter"`, or `"year"`)
- `periodValue`: The specific period value requested

**Overall Statistics (`overall`):**
- `totalTasks`: Total number of tasks assigned in the period
- `completedTasks`: Number of completed tasks
- `activeTasks`: Number of active (non-completed) tasks
- `backlogTasks`: Number of tasks in backlog/todo status
- `tasksInProgress`: Number of tasks currently in progress
- `tasksInReview`: Number of tasks in review status
- `pendingTasks`: Number of tasks in pending status
- `totalTimeSpent`: Total time spent in minutes
- `averageTimePerTask`: Average time spent per task in minutes
- `completionRate`: Completion rate as a percentage (0-100)

**By Project (`byProject`):**
Array of project-specific statistics, each containing:
- `projectId`: Project's unique identifier
- `projectName`: Project's name
- `stats`: Same structure as `overall` statistics, but filtered for this project

**By Time Period (`byTime`):**
Array of time period-specific statistics, each containing:
- `period`: Period identifier (e.g., `"2024-01"`, `"2024-Q1"`, `"2024"`)
- `periodLabel`: Human-readable period label (e.g., `"January 2024"`, `"Q1 2024"`, `"2024"`)
- `stats`: Same structure as `overall` statistics, but filtered for this time period

**Analysis (`analysis`):**
- `productivityTrend`: Trend indicator - `"increasing"`, `"decreasing"`, or `"stable"`
- `mostActiveProject`: Name of the project with the most tasks
- `mostActiveProjectId`: ID of the most active project
- `averageCompletionTime`: Average time to complete a task in minutes
- `peakProductivityMonth`: The period with the highest number of completed tasks
- `taskDistribution`: Map of task status to count

**Error Responses:**

**400 Bad Request:**
```json
{
  "message": "period parameter is required (month, quarter, or year)"
}
```

```json
{
  "message": "periodValue parameter is required"
}
```

```json
{
  "message": "Error fetching employee task stats",
  "error": "invalid month format. Expected YYYY-MM"
}
```

**404 Not Found:**
```json
{
  "message": "Employee not found"
}
```

**Examples:**

**Get monthly statistics:**
```bash
curl -X GET "https://api.example.com/api/employee/stats/user123?period=month&periodValue=2024-01" \
  -H "Authorization: Bearer your-token-here"
```

**Get quarterly statistics:**
```bash
curl -X GET "https://api.example.com/api/employee/stats/user123?period=quarter&periodValue=2024-Q1" \
  -H "Authorization: Bearer your-token-here"
```

**Get yearly statistics:**
```bash
curl -X GET "https://api.example.com/api/employee/stats/user123?period=year&periodValue=2024" \
  -H "Authorization: Bearer your-token-here"
```

**Get statistics for a specific project:**
```bash
curl -X GET "https://api.example.com/api/employee/stats/user123?period=month&periodValue=2024-01&projectId=proj456" \
  -H "Authorization: Bearer your-token-here"
```

---

## Task Status Values

The API recognizes the following task status values (case-insensitive):

- **Backlog/Todo:** `"backlog"`, `"todo"`, `"to-do"`
- **In Progress:** `"inprogress"`, `"in-progress"`, `"in_progress"`, `"in progress"`
- **In Review:** `"inreview"`, `"in-review"`, `"in_review"`, `"in review"`
- **Pending:** `"pending"`
- **Completed:** `"completed"`
- **Cancelled:** `"cancelled"`, `"canceled"`

## Notes

1. **Time Period Filtering:** The statistics API filters tasks based on their creation or last update date within the specified time period.

2. **Time Spent Calculation:** Time spent is calculated from the `TimeSpent` entries in tasks, filtered by the employee's user ID.

3. **Active Tasks:** Active tasks are defined as tasks that are not completed or cancelled.

4. **Completion Rate:** Calculated as `(completedTasks / totalTasks) * 100`.

5. **Productivity Trend:** Calculated by comparing task completion in the first half vs. second half of the time period.

6. **Performance:** The employee list API is optimized to fetch all data in a single pass, reducing database queries significantly.

## Rate Limiting

All endpoints are subject to rate limiting. Check response headers for rate limit information:
- `X-RateLimit-Limit`: Maximum number of requests allowed
- `X-RateLimit-Remaining`: Number of requests remaining
- `X-RateLimit-Reset`: Time when the rate limit resets

## Support

For issues or questions, please contact the API support team or refer to the main API documentation.

