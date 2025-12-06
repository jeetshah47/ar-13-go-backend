# Calendar API Documentation

This document describes the Calendar Event API endpoints and request/response formats, including Google Calendar and Google Meet integration.

## Endpoints

- `POST /api/calendar/add` - Create a new calendar event
- `GET /api/calendar/month/:year/:month` - Get events for a specific month
- `GET /api/calendar/event/:id` - Get a specific event by ID
- `PUT /api/calendar/update/:id` - Update an existing event
- `DELETE /api/calendar/delete/:id` - Delete an event

## Authentication

All endpoints require authentication via Bearer token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## Create Calendar Event

### Endpoint
```
POST /api/calendar/add
```

### Request Examples

#### 1. Offline Event (In-Person Meeting)

```json
{
  "title": "Team Standup Meeting",
  "category": "meeting",
  "priority": "high",
  "start": "2024-01-15T10:00:00Z",
  "end": "2024-01-15T11:00:00Z",
  "time": "10:00",
  "description": "Daily team standup meeting",
  "isRepeating": false,
  "createdBy": "user123",
  "addToGoogleCalendar": true,
  "eventType": "offline"
}
```

#### 2. Online Event with Google Meet (Using Duration)

```json
{
  "title": "Project Review Meeting",
  "category": "meeting",
  "priority": "high",
  "start": "2024-01-15T14:00:00Z",
  "description": "Quarterly project review and planning",
  "isRepeating": false,
  "createdBy": "user123",
  "addToGoogleCalendar": true,
  "eventType": "online",
  "invitedMemberIds": [
    "user456",
    "user789",
    "user101"
  ],
  "duration": 60
}
```

#### 3. Online Event with Google Meet (Using End Time)

```json
{
  "title": "Client Presentation",
  "category": "presentation",
  "priority": "high",
  "start": "2024-01-20T09:00:00Z",
  "end": "2024-01-20T10:30:00Z",
  "description": "Presenting Q1 results to client",
  "isRepeating": false,
  "createdBy": "user123",
  "addToGoogleCalendar": true,
  "eventType": "online",
  "invitedMemberIds": [
    "user456",
    "user789"
  ]
}
```

#### 4. Repeating Online Event (Weekly Meeting)

```json
{
  "title": "Weekly Team Sync",
  "category": "meeting",
  "priority": "medium",
  "start": "2024-01-15T15:00:00Z",
  "end": "2024-01-15T16:00:00Z",
  "description": "Weekly team synchronization meeting",
  "isRepeating": true,
  "repeatFrequency": "weekly",
  "repeatDays": ["monday"],
  "createdBy": "user123",
  "addToGoogleCalendar": true,
  "eventType": "online",
  "invitedMemberIds": [
    "user456",
    "user789",
    "user101"
  ],
  "duration": 60
}
```

#### 5. Simple Event (Minimal Fields)

```json
{
  "title": "Quick Check-in",
  "start": "2024-01-15T16:00:00Z",
  "end": "2024-01-15T16:30:00Z",
  "createdBy": "user123"
}
```

### Request Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | Yes | Event title |
| `category` | string | No | Event category (e.g., "meeting", "presentation", "task") |
| `priority` | string | No | Priority level: "high", "medium", "low" |
| `start` | string (ISO 8601) | Yes | Start time in UTC format (e.g., "2024-01-15T10:00:00Z") |
| `end` | string (ISO 8601) | No* | End time in UTC format. If not provided and `duration` is set, end time will be calculated automatically |
| `time` | string | No | Time in HH:MM format (e.g., "10:00") |
| `description` | string | No | Event description |
| `isRepeating` | boolean | No | Whether the event repeats (default: false) |
| `repeatFrequency` | string | No | Frequency of repetition: "daily", "weekly", "monthly" (required if `isRepeating` is true) |
| `repeatDays` | array[string] | No | Days of week for repeating events (e.g., ["monday", "wednesday"]) |
| `createdBy` | string | Yes | User ID of the event creator |
| `addToGoogleCalendar` | boolean | No | Whether to sync with Google Calendar (default: false) |
| `eventType` | string | No | Event type: "offline" or "online" (default: "offline") |
| `invitedMemberIds` | array[string] | No | Array of user IDs to invite (required for online events) |
| `duration` | integer | No* | Duration in minutes (used if `end` is not provided) |
| `googleMeetLink` | string | No | Auto-generated Google Meet link (returned in response for online events, read-only) |
| `googleCalendarEventId` | string | No | Google Calendar event ID (returned in response if synced, read-only) |

**Note:** Either `end` or `duration` should be provided. If both are provided, `end` takes precedence.

### Response Examples

#### Success Response (Online Event)

```json
{
  "message": "Calendar event added successfully",
  "event": {
    "id": "event123",
    "title": "Project Review Meeting",
    "category": "meeting",
    "priority": "high",
    "start": "2024-01-15T14:00:00Z",
    "end": "2024-01-15T15:00:00Z",
    "description": "Quarterly project review and planning",
    "isRepeating": false,
    "createdBy": "user123",
    "addToGoogleCalendar": true,
    "eventType": "online",
    "invitedMemberIds": ["user456", "user789", "user101"],
    "duration": 60,
    "googleMeetLink": "https://meet.google.com/abc-defg-hij",
    "googleCalendarEventId": "google_event_id_123",
    "created": "2024-01-10T10:00:00Z"
  }
}
```

#### Success Response (Offline Event)

```json
{
  "message": "Calendar event added successfully",
  "event": {
    "id": "event456",
    "title": "Team Standup Meeting",
    "category": "meeting",
    "priority": "high",
    "start": "2024-01-15T10:00:00Z",
    "end": "2024-01-15T11:00:00Z",
    "description": "Daily team standup meeting",
    "isRepeating": false,
    "createdBy": "user123",
    "addToGoogleCalendar": true,
    "eventType": "offline",
    "created": "2024-01-10T10:00:00Z"
  }
}
```

#### Error Response

```json
{
  "error": "Invalid month or year"
}
```

## Google Calendar Integration

### Features

1. **Automatic Sync**: When `addToGoogleCalendar` is set to `true`, events are automatically synced to Google Calendar
2. **Google Meet**: For online events (`eventType: "online"`), a Google Meet link is automatically created
3. **Attendees**: Invited members are automatically added as attendees in Google Calendar
4. **Timezone**: Events are converted to Indian Standard Time (IST - Asia/Kolkata) before syncing

### Prerequisites

1. **Google OAuth Setup**: Users must link their Google account via `/api/google-account/auth/initiate`
2. **Required Scopes**: The following OAuth scopes are required:
   - `https://www.googleapis.com/auth/calendar` (full calendar access)
   - `https://www.googleapis.com/auth/calendar.readonly` (read-only access)
   - `openid`
   - `email`
   - `profile`

### How It Works

1. **Creating an Online Event**:
   - Set `eventType: "online"`
   - Provide `invitedMemberIds` (array of user IDs)
   - Provide `duration` (in minutes) or `end` time
   - Set `addToGoogleCalendar: true`
   - The system will:
     - Create the event in your database
     - Create a Google Calendar event with Google Meet conference
     - Add invited members as attendees
     - Return the Google Meet link in the response

2. **Creating an Offline Event**:
   - Set `eventType: "offline"` or omit it
   - Set `addToGoogleCalendar: true` if you want to sync
   - No Google Meet link will be created

3. **Updating Events**:
   - Updates to synced events are automatically reflected in Google Calendar
   - Google Meet links persist across updates

4. **Deleting Events**:
   - Deleting synced events also removes them from Google Calendar

## Event Types

### Offline Events
- In-person meetings or events
- No Google Meet link created
- Can still be synced to Google Calendar

### Online Events
- Virtual meetings with Google Meet
- Requires `invitedMemberIds` array
- Google Meet link automatically generated
- Invited members receive calendar invites

## Repeating Events

### Supported Frequencies
- `daily` - Event repeats daily
- `weekly` - Event repeats weekly
- `monthly` - Event repeats monthly

### Repeat Days (for weekly events)
- `["monday"]` - Every Monday
- `["monday", "wednesday", "friday"]` - Every Monday, Wednesday, and Friday
- etc.

### Example: Daily Standup

```json
{
  "title": "Daily Standup",
  "start": "2024-01-15T09:00:00Z",
  "end": "2024-01-15T09:30:00Z",
  "isRepeating": true,
  "repeatFrequency": "daily",
  "createdBy": "user123"
}
```

### Example: Weekly Team Meeting

```json
{
  "title": "Weekly Team Meeting",
  "start": "2024-01-15T14:00:00Z",
  "end": "2024-01-15T15:00:00Z",
  "isRepeating": true,
  "repeatFrequency": "weekly",
  "repeatDays": ["monday"],
  "createdBy": "user123"
}
```

## Best Practices

1. **Always provide end time or duration** to ensure proper event scheduling
2. **Use ISO 8601 format** for date/time fields (UTC recommended)
3. **For online events**, always include `invitedMemberIds` to ensure attendees are notified
4. **Set `addToGoogleCalendar: true`** only if the user has linked their Google account
5. **Use appropriate event types** - "online" for virtual meetings, "offline" for in-person events

## Error Handling

### Common Errors

1. **400 Bad Request**: Invalid request format or missing required fields
   ```json
   {
     "error": "Invalid month or year"
   }
   ```

2. **401 Unauthorized**: Missing or invalid authentication token
   ```json
   {
     "error": "User not authenticated"
   }
   ```

3. **500 Internal Server Error**: Server-side error (e.g., Google Calendar sync failure)
   ```json
   {
     "error": "Failed to create Google Calendar event"
   }
   ```

## Notes

- All timestamps should be in UTC format (ISO 8601)
- Duration is specified in minutes
- Google Meet links are only generated for online events
- Invited members must have valid user IDs in the system
- Google Calendar sync happens asynchronously (non-blocking)
- Events are stored in Indian Standard Time (IST - Asia/Kolkata) when synced to Google Calendar

