# WebSocket Client Integration Guide

This guide provides step-by-step instructions for integrating the WebSocket service into your client application.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Installation](#installation)
3. [Basic Connection](#basic-connection)
4. [React Integration](#react-integration)
5. [Vue.js Integration](#vuejs-integration)
6. [Vanilla JavaScript](#vanilla-javascript)
7. [TypeScript Types](#typescript-types)
8. [Event Handling](#event-handling)
9. [Reconnection Strategy](#reconnection-strategy)
10. [Error Handling](#error-handling)
11. [Best Practices](#best-practices)
12. [Troubleshooting](#troubleshooting)

## Quick Start

```javascript
// 1. Get your JWT token (from your auth system)
const token = localStorage.getItem('authToken');

// 2. Connect to WebSocket
const ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);

// 3. Handle connection
ws.onopen = () => {
  console.log('Connected to WebSocket');
};

// 4. Listen for messages
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

// 5. Send messages
ws.send(JSON.stringify({
  type: 'task:update-status',
  data: {
    projectId: 'project-123',
    taskId: 'task-456',
    status: 'In Progress'
  }
}));
```

## Installation

### Browser (No Installation Required)

The native `WebSocket` API is built into all modern browsers. No installation needed!

### Node.js

```bash
npm install ws
# or
yarn add ws
```

```javascript
const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:3000/ws?token=your-token');
```

## Basic Connection

### Connection URL

```
Development: ws://localhost:3000/ws?token=YOUR_JWT_TOKEN
Production:  wss://your-domain.com/ws?token=YOUR_JWT_TOKEN
```

### Connection States

```javascript
const ws = new WebSocket('ws://localhost:3000/ws?token=your-token');

// Check connection state
console.log(ws.readyState);
// 0 = CONNECTING
// 1 = OPEN
// 2 = CLOSING
// 3 = CLOSED

// Wait for connection
ws.onopen = () => {
  console.log('WebSocket is open');
  // Now you can send messages
};
```

## React Integration

### Custom Hook Implementation

Create `hooks/useWebSocket.ts`:

```typescript
import { useEffect, useRef, useState, useCallback } from 'react';

interface Message {
  type: string;
  data: any;
}

interface UseWebSocketOptions {
  token: string | null;
  url?: string;
  onMessage?: (message: Message) => void;
  onError?: (error: Event) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  reconnect?: boolean;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

export function useWebSocket(options: UseWebSocketOptions) {
  const {
    token,
    url = 'ws://localhost:3000/ws',
    onMessage,
    onError,
    onConnect,
    onDisconnect,
    reconnect = true,
    reconnectInterval = 3000,
    maxReconnectAttempts = 5,
  } = options;

  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<Message | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const connect = useCallback(() => {
    if (!token) {
      console.warn('WebSocket: No token provided');
      return;
    }

    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close();
    }

    const ws = new WebSocket(`${url}?token=${token}`);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
      reconnectAttemptsRef.current = 0;
      onConnect?.();
    };

    ws.onmessage = (event) => {
      try {
        const message: Message = JSON.parse(event.data);
        setLastMessage(message);
        onMessage?.(message);
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      onError?.(error);
    };

    ws.onclose = (event) => {
      console.log('WebSocket disconnected:', event.code, event.reason);
      setIsConnected(false);
      onDisconnect?.();

      // Attempt reconnection
      if (reconnect && reconnectAttemptsRef.current < maxReconnectAttempts) {
        reconnectAttemptsRef.current++;
        console.log(
          `Reconnecting... (${reconnectAttemptsRef.current}/${maxReconnectAttempts})`
        );
        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, reconnectInterval * reconnectAttemptsRef.current); // Exponential backoff
      } else if (reconnectAttemptsRef.current >= maxReconnectAttempts) {
        console.error('Max reconnection attempts reached');
      }
    };
  }, [token, url, onMessage, onError, onConnect, onDisconnect, reconnect, reconnectInterval, maxReconnectAttempts]);

  const sendMessage = useCallback((type: string, data: any) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, data }));
    } else {
      console.warn('WebSocket is not open. Message not sent:', { type, data });
    }
  }, []);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  useEffect(() => {
    if (token) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [token, connect, disconnect]);

  return {
    isConnected,
    lastMessage,
    sendMessage,
    disconnect,
    reconnect: connect,
  };
}
```

### Usage in Component

```typescript
import React, { useEffect } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';
import { useAuth } from '../hooks/useAuth'; // Your auth hook

export function TaskBoard() {
  const { token } = useAuth();
  const { isConnected, lastMessage, sendMessage } = useWebSocket({
    token,
    onMessage: (message) => {
      console.log('Received message:', message);
    },
    onConnect: () => {
      console.log('Connected to WebSocket');
    },
    onDisconnect: () => {
      console.log('Disconnected from WebSocket');
    },
  });

  // Handle task status updates
  useEffect(() => {
    if (lastMessage?.type === 'task:status-updated') {
      const { projectId, taskId, status } = lastMessage.data;
      console.log(`Task ${taskId} status updated to ${status}`);
      // Update your UI state here
    }
  }, [lastMessage]);

  const handleTaskStatusUpdate = (
    projectId: string,
    taskId: string,
    status: string
  ) => {
    sendMessage('task:update-status', {
      projectId,
      taskId,
      status,
    });
  };

  return (
    <div>
      <div className="connection-status">
        Status: {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
      </div>
      {/* Your task board UI */}
      <button
        onClick={() =>
          handleTaskStatusUpdate('project-123', 'task-456', 'In Progress')
        }
      >
        Update Task Status
      </button>
    </div>
  );
}
```

## Vue.js Integration

### Composable Implementation

Create `composables/useWebSocket.ts`:

```typescript
import { ref, onMounted, onUnmounted, watch } from 'vue';

interface Message {
  type: string;
  data: any;
}

export function useWebSocket(token: string | null) {
  const isConnected = ref(false);
  const lastMessage = ref<Message | null>(null);
  let ws: WebSocket | null = null;
  let reconnectAttempts = 0;
  const maxReconnectAttempts = 5;

  const connect = () => {
    if (!token) {
      console.warn('WebSocket: No token provided');
      return;
    }

    if (ws) {
      ws.close();
    }

    ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);

    ws.onopen = () => {
      console.log('WebSocket connected');
      isConnected.value = true;
      reconnectAttempts = 0;
    };

    ws.onmessage = (event) => {
      try {
        const message: Message = JSON.parse(event.data);
        lastMessage.value = message;
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      isConnected.value = false;

      // Attempt reconnection
      if (reconnectAttempts < maxReconnectAttempts) {
        reconnectAttempts++;
        setTimeout(() => {
          connect();
        }, 3000 * reconnectAttempts);
      }
    };
  };

  const sendMessage = (type: string, data: any) => {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type, data }));
    }
  };

  const disconnect = () => {
    if (ws) {
      ws.close();
      ws = null;
    }
  };

  watch(token, (newToken) => {
    if (newToken) {
      connect();
    } else {
      disconnect();
    }
  });

  onMounted(() => {
    if (token) {
      connect();
    }
  });

  onUnmounted(() => {
    disconnect();
  });

  return {
    isConnected,
    lastMessage,
    sendMessage,
    disconnect,
  };
}
```

### Usage in Component

```vue
<template>
  <div>
    <div>Status: {{ isConnected ? 'Connected' : 'Disconnected' }}</div>
    <button @click="updateTaskStatus">Update Task</button>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue';
import { useWebSocket } from '@/composables/useWebSocket';
import { useAuth } from '@/composables/useAuth';

const { token } = useAuth();
const { isConnected, lastMessage, sendMessage } = useWebSocket(token);

watch(lastMessage, (message) => {
  if (message?.type === 'task:status-updated') {
    console.log('Task updated:', message.data);
  }
});

const updateTaskStatus = () => {
  sendMessage('task:update-status', {
    projectId: 'project-123',
    taskId: 'task-456',
    status: 'In Progress',
  });
};
</script>
```

## Vanilla JavaScript

### Class-Based Implementation

```javascript
class WebSocketClient {
  constructor(url, token) {
    this.url = url;
    this.token = token;
    this.ws = null;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
    this.messageHandlers = new Map();
    this.onConnectCallbacks = [];
    this.onDisconnectCallbacks = [];
    this.onErrorCallbacks = [];
  }

  connect() {
    if (!this.token) {
      console.error('WebSocket: No token provided');
      return;
    }

    this.ws = new WebSocket(`${this.url}?token=${this.token}`);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.isConnected = true;
      this.reconnectAttempts = 0;
      this.reconnectDelay = 1000;
      this.onConnectCallbacks.forEach((callback) => callback());
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('Error parsing message:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.onErrorCallbacks.forEach((callback) => callback(error));
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.isConnected = false;
      this.onDisconnectCallbacks.forEach((callback) => callback());
      this.attemptReconnect();
    };
  }

  handleMessage(message) {
    const { type, data } = message;

    // Call registered handlers
    if (this.messageHandlers.has(type)) {
      this.messageHandlers.get(type).forEach((handler) => handler(data));
    }

    // Call wildcard handler
    if (this.messageHandlers.has('*')) {
      this.messageHandlers.get('*').forEach((handler) => handler(message));
    }
  }

  on(type, handler) {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, []);
    }
    this.messageHandlers.get(type).push(handler);
  }

  off(type, handler) {
    if (this.messageHandlers.has(type)) {
      const handlers = this.messageHandlers.get(type);
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  onConnect(callback) {
    this.onConnectCallbacks.push(callback);
  }

  onDisconnect(callback) {
    this.onDisconnectCallbacks.push(callback);
  }

  onError(callback) {
    this.onErrorCallbacks.push(callback);
  }

  send(type, data) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }));
    } else {
      console.warn('WebSocket is not open');
    }
  }

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(
        `Reconnecting... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`
      );
      setTimeout(() => {
        this.connect();
      }, this.reconnectDelay);
      this.reconnectDelay *= 2; // Exponential backoff
    } else {
      console.error('Max reconnection attempts reached');
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

// Usage
const wsClient = new WebSocketClient('ws://localhost:3000/ws', 'your-token');

wsClient.on('authenticated', (data) => {
  console.log('Authenticated as:', data.userId);
});

wsClient.on('task:status-updated', (data) => {
  console.log('Task updated:', data);
});

wsClient.onConnect(() => {
  console.log('Connected!');
});

wsClient.connect();

// Send a message
wsClient.send('task:update-status', {
  projectId: 'project-123',
  taskId: 'task-456',
  status: 'In Progress',
});
```

## TypeScript Types

Create `types/websocket.ts`:

```typescript
// Message Types
export interface WebSocketMessage {
  type: string;
  data: any;
}

// Authentication
export interface AuthenticatedMessage {
  type: 'authenticated';
  data: {
    userId: string;
  };
}

// Task Status Update
export interface TaskUpdateStatusRequest {
  type: 'task:update-status';
  data: {
    projectId: string;
    taskId: string;
    status: string;
  };
}

export interface TaskUpdateStatusSuccess {
  type: 'task:update-status:success';
  data: {
    projectId: string;
    taskId: string;
    status: string;
    updatedBy: string;
    task?: any; // Full task object
  };
}

export interface TaskStatusUpdated {
  type: 'task:status-updated';
  data: {
    projectId: string;
    taskId: string;
    status: string;
    updatedBy: string;
    task?: any; // Full task object
  };
}

// Error
export interface ErrorMessage {
  type: 'error';
  data: {
    error: string;
  };
}

// Union type for all message types
export type WebSocketMessageType =
  | AuthenticatedMessage
  | TaskUpdateStatusRequest
  | TaskUpdateStatusSuccess
  | TaskStatusUpdated
  | ErrorMessage;
```

## Event Handling

### Available Events

#### Client-Sent Events

**`task:update-status`**
```typescript
ws.send(JSON.stringify({
  type: 'task:update-status',
  data: {
    projectId: 'project-123',
    taskId: 'task-456',
    status: 'In Progress' // or 'To Do', 'In Progress', 'Done', etc.
  }
}));
```

#### Server-Sent Events

**`authenticated`** - Sent immediately after connection
```typescript
{
  type: 'authenticated',
  data: {
    userId: 'user-id'
  }
}
```

**`task:update-status:success`** - Sent to the sender after successful update
```typescript
{
  type: 'task:update-status:success',
  data: {
    projectId: 'project-123',
    taskId: 'task-456',
    status: 'In Progress',
    updatedBy: 'user-id',
    task: { /* full task object */ }
  }
}
```

**`task:status-updated`** - Broadcasted to all project members
```typescript
{
  type: 'task:status-updated',
  data: {
    projectId: 'project-123',
    taskId: 'task-456',
    status: 'In Progress',
    updatedBy: 'user-id',
    task: { /* full task object */ }
  }
}
```

**`error`** - Sent when an error occurs
```typescript
{
  type: 'error',
  data: {
    error: 'Error message'
  }
}
```

## Reconnection Strategy

### Automatic Reconnection with Exponential Backoff

```javascript
class ReconnectingWebSocket {
  constructor(url, token, options = {}) {
    this.url = url;
    this.token = token;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = options.maxReconnectAttempts || 5;
    this.reconnectDelay = options.reconnectDelay || 1000;
    this.maxReconnectDelay = options.maxReconnectDelay || 30000;
    this.shouldReconnect = true;
  }

  connect() {
    this.ws = new WebSocket(`${this.url}?token=${this.token}`);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.reconnectDelay = 1000;
    };

    this.ws.onclose = () => {
      if (this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        const delay = Math.min(
          this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
          this.maxReconnectDelay
        );
        console.log(`Reconnecting in ${delay}ms... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        setTimeout(() => this.connect(), delay);
      }
    };
  }

  disconnect() {
    this.shouldReconnect = false;
    if (this.ws) {
      this.ws.close();
    }
  }
}
```

## Error Handling

### Comprehensive Error Handling

```javascript
const ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
  // Handle error (show notification, log, etc.)
};

ws.onclose = (event) => {
  if (event.code === 1006) {
    // Abnormal closure - network error
    console.error('Connection closed abnormally');
  } else if (event.code === 1008) {
    // Policy violation
    console.error('Connection closed due to policy violation');
  } else if (event.code === 1002) {
    // Protocol error
    console.error('Protocol error');
  }
  // Implement reconnection logic
};

// Handle message errors
ws.onmessage = (event) => {
  try {
    const message = JSON.parse(event.data);
    if (message.type === 'error') {
      console.error('Server error:', message.data.error);
      // Handle server-side errors
    }
  } catch (error) {
    console.error('Error parsing message:', error);
  }
};
```

## Best Practices

### 1. Token Management

```javascript
// Always check token validity before connecting
function connectWebSocket() {
  const token = getAuthToken(); // Your token getter
  
  if (!token || isTokenExpired(token)) {
    // Refresh token first
    refreshToken().then((newToken) => {
      connectWithToken(newToken);
    });
  } else {
    connectWithToken(token);
  }
}
```

### 2. Connection State Management

```javascript
// Use a state machine for connection management
const ConnectionState = {
  DISCONNECTED: 'disconnected',
  CONNECTING: 'connecting',
  CONNECTED: 'connected',
  RECONNECTING: 'reconnecting',
  ERROR: 'error',
};

let connectionState = ConnectionState.DISCONNECTED;
```

### 3. Message Queue

```javascript
// Queue messages when disconnected
class MessageQueue {
  constructor() {
    this.queue = [];
  }

  add(type, data) {
    this.queue.push({ type, data, timestamp: Date.now() });
  }

  flush(ws) {
    if (ws.readyState === WebSocket.OPEN) {
      while (this.queue.length > 0) {
        const message = this.queue.shift();
        ws.send(JSON.stringify(message));
      }
    }
  }
}

const messageQueue = new MessageQueue();

ws.onopen = () => {
  messageQueue.flush(ws);
};
```

### 4. Cleanup on Unmount

```javascript
// Always cleanup WebSocket connections
useEffect(() => {
  const ws = new WebSocket(url);
  
  return () => {
    ws.close();
  };
}, []);
```

### 5. Rate Limiting

```javascript
// Implement rate limiting for messages
class RateLimiter {
  constructor(maxMessages, windowMs) {
    this.maxMessages = maxMessages;
    this.windowMs = windowMs;
    this.messages = [];
  }

  canSend() {
    const now = Date.now();
    this.messages = this.messages.filter(
      (timestamp) => now - timestamp < this.windowMs
    );
    return this.messages.length < this.maxMessages;
  }

  record() {
    this.messages.push(Date.now());
  }
}

const rateLimiter = new RateLimiter(10, 1000); // 10 messages per second

function sendMessage(type, data) {
  if (rateLimiter.canSend()) {
    ws.send(JSON.stringify({ type, data }));
    rateLimiter.record();
  } else {
    console.warn('Rate limit exceeded');
  }
}
```

## Troubleshooting

### Connection Issues

**Problem: Connection fails immediately**
- ✅ Check if token is valid and not expired
- ✅ Verify WebSocket URL is correct
- ✅ Check CORS settings on server
- ✅ Check browser console for errors
- ✅ Verify server is running

**Problem: Connection closes unexpectedly**
- ✅ Check network stability
- ✅ Verify token hasn't expired
- ✅ Check server logs for errors
- ✅ Verify ping/pong is working (handled automatically by browser)

**Problem: Messages not received**
- ✅ Check `ws.readyState === WebSocket.OPEN`
- ✅ Verify message format is valid JSON
- ✅ Check event handler is properly set up
- ✅ Verify message type matches expected format

### Common Errors

**401 Unauthorized**
```
Solution: Token is missing, invalid, or expired. Refresh your token.
```

**WebSocket is closed before connection is established**
```
Solution: Check token validity and server availability.
```

**Invalid message format**
```
Solution: Ensure all messages follow the { type: string, data: any } format.
```

## Example: Complete Integration

```typescript
// Complete example with all features
import { useEffect, useRef, useState } from 'react';

interface WebSocketConfig {
  token: string | null;
  url?: string;
}

export function useCompleteWebSocket(config: WebSocketConfig) {
  const { token, url = 'ws://localhost:3000/ws' } = config;
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<any>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const messageQueueRef = useRef<Array<{ type: string; data: any }>>([]);

  const connect = () => {
    if (!token) return;

    if (wsRef.current) {
      wsRef.current.close();
    }

    const ws = new WebSocket(`${url}?token=${token}`);
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
      // Flush queued messages
      messageQueueRef.current.forEach((msg) => {
        ws.send(JSON.stringify(msg));
      });
      messageQueueRef.current = [];
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        setLastMessage(message);
      } catch (error) {
        console.error('Error parsing message:', error);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
      setIsConnected(false);
      // Reconnect after 3 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        connect();
      }, 3000);
    };
  };

  const sendMessage = (type: string, data: any) => {
    const message = { type, data };
    
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    } else {
      // Queue message for later
      messageQueueRef.current.push(message);
    }
  };

  useEffect(() => {
    if (token) {
      connect();
    }

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [token]);

  return {
    isConnected,
    lastMessage,
    sendMessage,
  };
}
```

## Additional Resources

- [WebSocket API Documentation](./WEBSOCKET_API.md) - Full API reference
- [MDN WebSocket Documentation](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
- [React Hooks Documentation](https://react.dev/reference/react)

---

**Need Help?** Check the [Troubleshooting](#troubleshooting) section or review server logs for detailed error messages.

