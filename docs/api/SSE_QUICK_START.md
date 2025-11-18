# SSE Quick Start Guide

Quick reference for integrating SSE into your client application.

## Basic Connection

```javascript
const token = 'your-jwt-token';
const eventSource = new EventSource(`/api/events?token=${token}`);

eventSource.addEventListener('authenticated', (event) => {
  const data = JSON.parse(event.data);
  console.log('Connected as:', data.userId);
});

eventSource.addEventListener('task:status-updated', (event) => {
  const data = JSON.parse(event.data);
  // Update your UI here
  updateTask(data.task);
});
```

## Update Task Status

```javascript
async function updateTaskStatus(projectId, taskId, status) {
  const response = await fetch(`/api/tasks/update-status/${projectId}/${taskId}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ status })
  });
  
  return response.json();
}
```

## React Hook (Copy-Paste Ready)

```typescript
import { useEffect, useRef, useState } from 'react';

export function useSSE(token: string, baseURL: string = '') {
  const [isConnected, setIsConnected] = useState(false);
  const [lastEvent, setLastEvent] = useState<any>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!token) return;

    const eventSource = new EventSource(`${baseURL}/api/events?token=${token}`);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => setIsConnected(true);
    eventSource.onerror = () => setIsConnected(false);

    eventSource.addEventListener('authenticated', (e) => {
      setLastEvent({ type: 'authenticated', data: JSON.parse(e.data) });
    });

    eventSource.addEventListener('task:status-updated', (e) => {
      setLastEvent({ type: 'task:status-updated', data: JSON.parse(e.data) });
    });

    return () => eventSource.close();
  }, [token, baseURL]);

  return { isConnected, lastEvent };
}
```

## Event Types

- `authenticated` - Connection established
- `task:status-updated` - Task status changed (broadcast)
- `task:update-status:success` - Your update succeeded
- `error` - Error occurred

## Valid Statuses

`pending`, `in_progress`, `in_review`, `completed`, `accepted`, `rejected`

