# Product Requirements Document (PRD)
## AR-13 Project Management & Operations Dashboard

**Document Version:** 1.0  
**Date:** January 2025  
**Status:** Active

---

## 1. Product Overview

### 1.1 Product Vision
AR-13 is a comprehensive, modern project management and operations dashboard that empowers teams to collaborate effectively, track work efficiently, and make data-driven decisions. The platform provides a unified interface for managing projects, tasks, team members, calendars, and organizational activities.

### 1.2 Product Mission
To provide organizations with a scalable, user-friendly platform that centralizes project management, enhances team productivity, and delivers actionable insights through real-time collaboration and comprehensive analytics.

### 1.3 Product Goals
1. **Unified Platform**: Single source of truth for all projects and tasks
2. **Real-Time Collaboration**: Instant updates and notifications for team coordination
3. **Data-Driven Insights**: Analytics and reporting for informed decision-making
4. **Seamless Integration**: Google Calendar and other third-party integrations
5. **Scalable Architecture**: Support growing teams and increasing complexity
6. **User Experience**: Intuitive interface requiring minimal training

### 1.4 Target Users
- **Primary**: Project managers, team leads, individual contributors
- **Secondary**: Executives, administrators, HR personnel
- **User Personas**:
  - **Project Manager**: Manages multiple projects, assigns tasks, tracks progress
  - **Team Member**: Receives tasks, updates progress, attends meetings
  - **Administrator**: Manages users, configures system, views analytics
  - **Executive**: Reviews high-level dashboards and organizational metrics

---

## 2. Product Architecture

### 2.1 System Architecture
- **Frontend**: React + TypeScript + Vite (ar-13-ui)
- **Backend**: Go (Golang) REST API (ar-13-go-backend)
- **Database**: AWS DynamoDB
- **Real-Time**: WebSocket for notifications
- **Authentication**: JWT-based authentication
- **Deployment**: Web application + Electron desktop app

### 2.2 Technology Stack

#### Frontend
- React 18.3+
- TypeScript 5.8+
- Vite 6.3+
- Material-UI (MUI) 7.1+
- Redux Toolkit for state management
- React Router for navigation
- Axios for API calls
- WebSocket client for real-time updates

#### Backend
- Go 1.21+
- Gin web framework
- AWS DynamoDB SDK
- JWT authentication
- WebSocket (Gorilla WebSocket)
- File upload handling
- Email service integration

### 2.3 Key Integrations
- Google Calendar API
- Google Meet API
- Google OAuth 2.0
- AWS Services (DynamoDB, S3 for file storage)

---

## 3. Functional Requirements

### 3.1 Authentication & Authorization

#### 3.1.1 User Authentication
**PRD-FR-001**: User Login
- Users can log in with email and password
- System generates JWT tokens for authenticated sessions
- Tokens expire after 24 hours (configurable)
- Refresh tokens valid for 30 days

**PRD-FR-002**: Google OAuth Integration
- Users can link Google accounts for calendar integration
- OAuth flow with Google Cloud Console
- Secure token storage and management

**PRD-FR-003**: Session Management
- Automatic token refresh
- Secure logout functionality
- Session timeout handling

#### 3.1.2 Role-Based Access Control (RBAC)
**PRD-FR-004**: User Roles
- **Admin Role**: Full system access
  - All CRUD operations on projects, tasks, users
  - User management and role assignment
  - System configuration
  - Analytics and reporting
- **Standard Role**: Limited access
  - Read access to assigned projects and tasks
  - Write access to assigned tasks
  - Calendar management
  - Profile management

**PRD-FR-005**: Permission System
- Granular permissions (e.g., `projects:read`, `tasks:write`, `users:delete`)
- Permission checking on all API endpoints
- Dynamic permission retrieval via `/api/auth/permissions`

### 3.2 Project Management

#### 3.2.1 Project CRUD Operations
**PRD-FR-006**: Create Project
- Project title, description, deadline
- Project owner assignment
- Initial member assignment
- Project creation activity log

**PRD-FR-007**: View Projects
- List all projects (filtered by user access)
- Project details view
- Project member list
- Project task summary

**PRD-FR-008**: Update Project
- Edit project details
- Add/remove project members
- Update project owner
- Update project deadline
- Activity logging for all changes

**PRD-FR-009**: Delete Project
- Soft delete or hard delete (configurable)
- Validation: cannot delete if active tasks exist
- Activity log entry

#### 3.2.2 Project Features
**PRD-FR-010**: Project Members
- Add members to projects
- Remove members from projects
- View member list with roles
- Member activity tracking

**PRD-FR-011**: Project Analytics
- Task count by status
- Completion percentage
- Member workload distribution
- Deadline tracking

### 3.3 Task Management

#### 3.3.1 Task CRUD Operations
**PRD-FR-012**: Create Task
- Task subject, description, code
- Project assignment (required)
- Task assignment to user(s)
- Priority (High, Medium, Low)
- Deadline (RFC3339 format)
- Initial status (Backlog, To-Do)
- Task creation activity log

**PRD-FR-013**: View Tasks
- List tasks by project
- Filter by status, assignee, priority
- Search tasks by subject/description
- Task detail view with full information
- Task activity log history

**PRD-FR-014**: Update Task
- Update task details (subject, description, priority)
- Update task status
- Update task deadline via `/api/tasks/update-deadline/:projectId/:taskId`
- Update task progress (0-100%) via `/api/tasks/update-progress/:projectId/:taskId`
- Reassign tasks
- Activity logging for all changes

**PRD-FR-015**: Delete Task
- Delete task (with validation)
- Activity log entry
- Cascade delete of related data (files, time entries)

#### 3.3.2 Task Features
**PRD-FR-016**: Task Status Management
- Status values: Backlog, To-Do, In Progress, In Review, Pending, Completed, Cancelled
- Status transition validation
- Automatic progress updates based on status
- Status change notifications

**PRD-FR-017**: Task Assignment
- Assign task to single or multiple users
- Unassign users from tasks
- Assignment notifications
- Assignment activity logging

**PRD-FR-018**: Task Progress Tracking
- Progress percentage (0-100)
- Progress update API endpoint
- Progress history tracking
- Progress-based analytics

**PRD-FR-019**: Task Deadline Management
- Deadline setting and updates
- Deadline validation (future dates)
- Deadline notifications
- Overdue task identification

**PRD-FR-020**: Time Tracking
- Log time spent on tasks
- Time entries with date, duration, description
- Update and delete time entries
- Time tracking analytics

**PRD-FR-021**: File Attachments
- Upload files to tasks
- File type validation
- File size limits (configurable)
- File download functionality
- File removal with activity logging

### 3.4 Calendar & Event Management

#### 3.4.1 Calendar Operations
**PRD-FR-022**: View Calendar
- Month view with events
- Event list by date
- Event filtering and search
- Calendar navigation

**PRD-FR-023**: Create Events
- Event title, description, category
- Start and end times (ISO 8601 format)
- Event type: Offline or Online
- Priority level
- Repeating events (daily, weekly, monthly)
- Google Calendar sync option
- Event invitations for online events

**PRD-FR-024**: Update Events
- Modify event details
- Reschedule events
- Update attendees
- Sync changes to Google Calendar

**PRD-FR-025**: Delete Events
- Delete single events
- Delete repeating event series
- Remove from Google Calendar if synced

#### 3.4.2 Google Calendar Integration
**PRD-FR-026**: Google Calendar Sync
- Automatic sync when `addToGoogleCalendar: true`
- Two-way sync (create, update, delete)
- Timezone conversion (IST - Asia/Kolkata)
- Google Calendar event ID tracking

**PRD-FR-027**: Google Meet Integration
- Automatic Google Meet link generation for online events
- Meet link included in calendar invites
- Meet link persistence across updates
- Attendee management

**PRD-FR-028**: Repeating Events
- Daily, weekly, monthly frequencies
- Day selection for weekly events
- Recurrence rule generation
- Google Calendar recurrence sync

### 3.5 User Management

#### 3.5.1 User Operations
**PRD-FR-029**: User Profile
- View and edit profile (name, email, phone, designation)
- Profile picture (future enhancement)
- Password management
- Profile update activity logging

**PRD-FR-030**: User List
- List all users (admin only)
- Filter and search users
- User role display
- User status (active/inactive)

**PRD-FR-031**: User Invitation
- Invite new users via email
- Invitation link generation
- User registration flow
- Role assignment during invitation

### 3.6 Employee Analytics & Reporting

#### 3.6.1 Employee Statistics
**PRD-FR-032**: Employee Task Counts
- Task counts by status (backlog, in progress, in review, pending)
- Total tasks and active tasks
- Employee list with aggregated counts
- Individual employee task counts API

**PRD-FR-033**: Employee Analytics
- Time period filtering (month, quarter, year)
- Overall statistics (total, completed, active tasks)
- Project-based statistics
- Time-based breakdown
- Productivity trends
- Completion rates
- Average time per task

**PRD-FR-034**: Analytics Features
- Most active project identification
- Peak productivity period analysis
- Task distribution by status
- Time spent analytics

### 3.7 Activity Logging & Audit

#### 3.7.1 Activity Log System
**PRD-FR-035**: Activity Logging
- Automatic logging of all entity operations
- Entity types: Tasks, Projects, Users, Calendar Events
- Action types: created, updated, deleted, assigned, status_changed, etc.
- Field-level change tracking (old/new values)
- User attribution (who performed action)
- Timestamp tracking

**PRD-FR-036**: Activity Log Retrieval
- Get logs by entity (task, project, user, event)
- Get logs by entity type
- Get logs by user (who performed action)
- Log filtering and pagination
- Activity log API endpoints

**PRD-FR-037**: Audit Trail
- Immutable log entries (no modification/deletion)
- Comprehensive change history
- User activity tracking
- Compliance-ready audit trail

### 3.8 Drawing List Management

#### 3.8.1 Drawing Categories
**PRD-FR-038**: Category Management
- Create, read, update, delete drawing categories
- Category ordering
- Category activation/deactivation
- Default categories: Architectural, Structural, MEP, Site Plans

#### 3.8.2 Drawing Types
**PRD-FR-039**: Type Management
- Create, read, update, delete drawing types
- Type association with categories
- Type ordering within categories
- Default types per category
- Cannot delete category with associated types

**PRD-FR-040**: Drawing List API
- CRUD operations for categories and types
- Get categories with nested types
- Seeding script for initial data
- Permission-based access (`drawingList:read`, `drawingList:write`, `drawingList:delete`)

### 3.9 Real-Time Notifications

#### 3.9.1 WebSocket Integration
**PRD-FR-041**: WebSocket Connection
- JWT-authenticated WebSocket connections
- Single active connection per user
- Automatic reconnection logic
- Keep-alive ping/pong mechanism

**PRD-FR-042**: Real-Time Updates
- Task status change notifications
- Task assignment notifications
- Project update notifications
- Calendar event notifications
- User activity notifications

**PRD-FR-043**: Notification Types
- Task updates (status, assignment, deadline)
- Project updates (members, deadline)
- Calendar event reminders
- System announcements

### 3.10 File Management

#### 3.10.1 File Operations
**PRD-FR-044**: File Upload
- Upload files to tasks
- File type validation
- File size limits
- File metadata storage
- Upload progress tracking

**PRD-FR-045**: File Management
- List files attached to tasks
- Download files
- Delete files
- File activity logging

### 3.11 Vacation/Leave Management

#### 3.11.1 Leave Operations
**PRD-FR-046**: Leave Requests
- Create leave requests
- View leave calendar
- Leave approval workflow (admin)
- Leave status tracking

**PRD-FR-047**: Leave Features
- Leave type categorization
- Leave duration calculation
- Leave balance tracking
- Leave calendar integration

### 3.12 Dashboard & Analytics

#### 3.12.1 Dashboard Views
**PRD-FR-048**: Project Dashboard
- Project overview cards
- Task status distribution
- Recent activity feed
- Upcoming deadlines

**PRD-FR-049**: Employee Dashboard
- Personal task list
- Task statistics
- Calendar view
- Recent notifications

**PRD-FR-050**: Executive Dashboard
- Organization-wide metrics
- Project portfolio view
- Resource utilization
- Productivity trends

---

## 4. Non-Functional Requirements

### 4.1 Performance
- **API Response Time**: 95th percentile under 2 seconds
- **Page Load Time**: Initial load under 3 seconds
- **WebSocket Latency**: Real-time updates within 500ms
- **Concurrent Users**: Support 500+ concurrent users
- **Database Queries**: Optimized queries with proper indexing

### 4.2 Security
- **Authentication**: JWT-based with secure token storage
- **Authorization**: Role-based access control on all endpoints
- **Data Encryption**: HTTPS/WSS for all communications
- **Input Validation**: All user inputs validated and sanitized
- **SQL Injection Prevention**: Parameterized queries
- **XSS Prevention**: Input sanitization and output encoding
- **CSRF Protection**: Token-based CSRF protection

### 4.3 Scalability
- **Horizontal Scaling**: Backend services support load balancing
- **Database Scaling**: DynamoDB auto-scaling configuration
- **Caching**: Redis caching for frequently accessed data (optional)
- **CDN**: Static asset delivery via CDN

### 4.4 Reliability
- **Uptime**: 99.5% availability during business hours
- **Error Handling**: Graceful error handling with user-friendly messages
- **Data Backup**: Automated daily backups
- **Disaster Recovery**: Recovery procedures documented
- **Monitoring**: Application and infrastructure monitoring

### 4.5 Usability
- **User Interface**: Modern, intuitive Material-UI design
- **Responsive Design**: Works on desktop, tablet, and mobile browsers
- **Accessibility**: WCAG 2.1 AA compliance (target)
- **Browser Support**: Chrome, Firefox, Safari, Edge (latest 2 versions)
- **Desktop App**: Electron-based desktop application

### 4.6 Maintainability
- **Code Quality**: TypeScript/Go type safety, code reviews
- **Documentation**: Comprehensive API and code documentation
- **Testing**: Unit tests, integration tests, E2E tests
- **Logging**: Structured logging for debugging and monitoring
- **Version Control**: Git-based version control with branching strategy

---

## 5. User Stories

### 5.1 Project Management
- **US-001**: As a project manager, I want to create a new project so that I can organize related tasks.
- **US-002**: As a team member, I want to view all projects I'm assigned to so that I know what I'm working on.
- **US-003**: As a project owner, I want to add team members to my project so that they can collaborate.

### 5.2 Task Management
- **US-004**: As a project manager, I want to create and assign tasks so that work is distributed.
- **US-005**: As a team member, I want to update my task status so that others know my progress.
- **US-006**: As a team member, I want to see my task deadlines so that I can prioritize work.
- **US-007**: As a project manager, I want to track task progress so that I can identify bottlenecks.

### 5.3 Calendar
- **US-008**: As a team member, I want to schedule meetings so that we can coordinate.
- **US-009**: As a team member, I want to sync events with Google Calendar so that I have a unified calendar.
- **US-010**: As a team member, I want to create online meetings with Google Meet so that we can collaborate remotely.

### 5.4 Analytics
- **US-011**: As an executive, I want to view employee productivity metrics so that I can make resource decisions.
- **US-012**: As a project manager, I want to see project completion rates so that I can track progress.

### 5.5 Notifications
- **US-013**: As a team member, I want to receive real-time notifications so that I'm aware of important updates.
- **US-014**: As a team member, I want to be notified when tasks are assigned to me so that I can start working.

---

## 6. API Specifications

### 6.1 Authentication APIs
- `POST /api/auth/login` - User login
- `POST /api/auth/register` - User registration
- `GET /api/auth/permissions` - Get user permissions
- `POST /api/auth/refresh` - Refresh JWT token

### 6.2 Project APIs
- `GET /api/projects` - List projects
- `POST /api/projects` - Create project
- `GET /api/projects/:id` - Get project details
- `PUT /api/projects/:id` - Update project
- `DELETE /api/projects/:id` - Delete project

### 6.3 Task APIs
- `GET /api/tasks/:projectId` - List tasks in project
- `POST /api/tasks/add` - Create task
- `GET /api/tasks/:projectId/:taskId` - Get task details
- `PUT /api/tasks/:projectId/:taskId` - Update task
- `PUT /api/tasks/update-deadline/:projectId/:taskId` - Update task deadline
- `PUT /api/tasks/update-progress/:projectId/:taskId` - Update task progress
- `DELETE /api/tasks/:projectId/:taskId` - Delete task

### 6.4 Calendar APIs
- `GET /api/calendar/month/:year/:month` - Get month events
- `POST /api/calendar/add` - Create event
- `GET /api/calendar/event/:id` - Get event details
- `PUT /api/calendar/update/:id` - Update event
- `DELETE /api/calendar/delete/:id` - Delete event

### 6.5 Employee APIs
- `GET /api/employee/list` - List employees with task counts
- `GET /api/employee/task-counts/:userId` - Get employee task counts
- `GET /api/employee/stats/:userId` - Get employee statistics

### 6.6 Activity Log APIs
- `GET /api/activity-logs/entity-types` - Get supported entity types
- `GET /api/activity-logs/:entityType/:entityId` - Get logs for entity
- `GET /api/activity-logs/:entityType` - Get logs by entity type

### 6.7 Drawing List APIs
- `GET /api/drawing-list/categories` - List categories
- `POST /api/drawing-list/categories` - Create category
- `GET /api/drawing-list/types` - List types
- `POST /api/drawing-list/types` - Create type
- `GET /api/drawing-list/categories-with-types` - Get categories with nested types

### 6.8 WebSocket API
- `WS /ws?token=<jwt>` - WebSocket connection for real-time updates

---

## 7. Data Models

### 7.1 Project Model
```typescript
{
  id: string;
  title: string;
  description?: string;
  ownerId: string;
  memberIds: string[];
  deadline?: string; // ISO 8601
  status: string;
  created: string; // ISO 8601
  updated?: string; // ISO 8601
}
```

### 7.2 Task Model
```typescript
{
  id: string;
  projectId: string;
  subject: string;
  code?: string;
  description?: string;
  status: string; // backlog, todo, in-progress, in-review, pending, completed, cancelled
  priority: string; // high, medium, low
  deadline?: string; // RFC3339
  progress?: number; // 0-100
  assignTo: string[];
  timeSpent: TimeEntry[];
  fileAttachments: FileAttachment[];
  created: string;
  updated?: string;
}
```

### 7.3 Calendar Event Model
```typescript
{
  id: string;
  title: string;
  category?: string;
  priority?: string;
  start: string; // ISO 8601
  end: string; // ISO 8601
  description?: string;
  isRepeating: boolean;
  repeatFrequency?: string; // daily, weekly, monthly
  repeatDays?: string[];
  createdBy: string;
  addToGoogleCalendar: boolean;
  eventType: string; // offline, online
  invitedMemberIds?: string[];
  googleMeetLink?: string;
  googleCalendarEventId?: string;
  created: string;
  updated?: string;
}
```

### 7.4 User Model
```typescript
{
  id: string;
  name: string;
  email: string;
  phoneNumber?: string;
  role: string; // Admin, Standard
  designation?: string;
  created: string;
  updated?: string;
}
```

### 7.5 Activity Log Model
```typescript
{
  id: string;
  entityType: string; // task, project, user, calendarEvent
  entityId: string;
  action: string; // created, updated, deleted, assigned, etc.
  createdAt: string;
  createdBy: string;
  fields?: Record<string, {old: any, new: any}>;
  description: string;
  metadata?: Record<string, any>;
}
```

---

## 8. User Interface Requirements

### 8.1 Design Principles
- **Modern & Clean**: Material-UI design system
- **Responsive**: Mobile-first responsive design
- **Accessible**: WCAG 2.1 AA compliance target
- **Consistent**: Unified design language across all pages
- **Intuitive**: Minimal learning curve

### 8.2 Key Pages/Screens
1. **Login/Registration Page**
2. **Dashboard** (Project overview, recent activity)
3. **Projects List Page**
4. **Project Detail Page** (Tasks, members, analytics)
5. **Task List Page** (Filtered by project/status)
6. **Task Detail Page** (Full task information, activity log)
7. **Calendar Page** (Month view, event list)
8. **Employee Analytics Page**
9. **User Profile Page**
10. **Admin Panel** (User management, system settings)

### 8.3 Key UI Components
- Navigation drawer/sidebar
- Data tables with sorting/filtering
- Forms with validation
- Modals and dialogs
- Toast notifications
- Loading indicators
- Charts and graphs (Recharts)
- Rich text editor (TipTap)

---

## 9. Integration Requirements

### 9.1 Google Calendar Integration
- OAuth 2.0 authentication flow
- Calendar event creation/update/deletion
- Google Meet link generation
- Attendee management
- Timezone handling (IST conversion)

### 9.2 Email Integration
- SMTP email service
- User invitation emails
- Notification emails (optional)
- Password reset emails

### 9.3 File Storage Integration
- AWS S3 or local file storage
- File upload/download APIs
- File metadata management

---

## 10. Testing Requirements

### 10.1 Unit Testing
- Frontend: Vitest for React components
- Backend: Go testing framework
- Target: 80% code coverage

### 10.2 Integration Testing
- API endpoint testing
- Database integration testing
- Third-party service mocking

### 10.3 End-to-End Testing
- Playwright for E2E tests
- Critical user flows
- Cross-browser testing

### 10.4 Performance Testing
- Load testing (500+ concurrent users)
- Stress testing
- API response time monitoring

---

## 11. Deployment & DevOps

### 11.1 Deployment Strategy
- **Frontend**: Static build deployment (Nginx, S3+CloudFront, or similar)
- **Backend**: Go binary on EC2 or containerized deployment
- **Database**: AWS DynamoDB
- **Desktop App**: Electron build for Windows, macOS, Linux

### 11.2 Environment Management
- Development environment
- Staging environment
- Production environment
- Environment-specific configuration

### 11.3 CI/CD Pipeline
- Automated testing on pull requests
- Automated builds
- Staging deployment on merge to main
- Production deployment with approval

---

## 12. Success Criteria

### 12.1 Launch Criteria
- All critical features implemented and tested
- Security audit completed
- Performance benchmarks met
- User documentation available
- Training materials prepared

### 12.2 Post-Launch Metrics
- User adoption rate (target: 90% within 6 months)
- Daily active users (target: 70% of registered users)
- Task completion rate improvement (target: 25%)
- User satisfaction score (target: 4.0+/5.0)
- System uptime (target: 99.5%)

---

## 13. Future Enhancements (Out of Scope for v1.0)

- Mobile native applications (iOS/Android)
- Advanced reporting and BI dashboards
- Custom workflow automation
- Advanced document collaboration
- Multi-tenant SaaS capabilities
- Advanced search and filtering
- Customizable dashboards
- API webhooks for third-party integrations
- Advanced notification preferences
- Time tracking with detailed reports

---

## 14. Appendix

### 14.1 Glossary
- **RBAC**: Role-Based Access Control
- **JWT**: JSON Web Token
- **API**: Application Programming Interface
- **WebSocket**: Real-time bidirectional communication protocol
- **DynamoDB**: AWS NoSQL database service
- **Electron**: Framework for building desktop applications

### 14.2 References
- Business Requirements Document (BRD)
- API Documentation
- Technical Architecture Document
- User Research Findings

---

**Document End**
