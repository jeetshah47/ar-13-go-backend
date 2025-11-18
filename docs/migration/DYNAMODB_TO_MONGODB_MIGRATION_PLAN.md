# DynamoDB to MongoDB Migration Plan

## Executive Summary

This document outlines a comprehensive plan to migrate the application from AWS DynamoDB to MongoDB. The migration will be executed in phases to minimize risk and ensure zero downtime.

**Migration Rationale:**
- Better query flexibility and complex queries
- More cost-effective for relational data patterns
- Easier schema evolution
- Better support for nested documents and arrays
- Simplified development experience

---

## Table of Contents

1. [Current State Analysis](#current-state-analysis)
2. [MongoDB Schema Design](#mongodb-schema-design)
3. [Migration Strategy](#migration-strategy)
4. [Implementation Phases](#implementation-phases)
5. [Data Migration Process](#data-migration-process)
6. [Code Changes Required](#code-changes-required)
7. [Testing Strategy](#testing-strategy)
8. [Rollback Plan](#rollback-plan)
9. [Timeline and Dependencies](#timeline-and-dependencies)
10. [Risk Assessment](#risk-assessment)

---

## Current State Analysis

### DynamoDB Tables

The application currently uses **12 DynamoDB tables**:

1. **users** - User accounts with email GSI
2. **projects** - Project information
3. **tasks** - Tasks with projectId GSI
4. **notifications** - User notifications with userId GSI
5. **calendar_events** - Calendar events
6. **leaveRequests** - Leave/vacation requests with userId and status GSIs
7. **activity_logs** - Activity logs with entityId GSI
8. **info-portal** - Info portal content
9. **project_details** - Project details with projectId GSI
10. **user_account_links** - User account links with userId GSI
11. **signupInvitations** - Signup invitations with email and token GSIs
12. **role_permissions** - Role-based permissions with role GSI

### Current Architecture

- **Repository Pattern**: All data access goes through repositories (`internal/repos/`)
- **Base Repository**: `DynamoBaseRepo` provides common CRUD operations
- **Service Layer**: Business logic in `internal/services/` uses repositories
- **Models**: Defined in `internal/models/` with DynamoDB tags
- **Dependencies**: AWS SDK v2 for DynamoDB operations

### Key Operations

- **GetByID**: Primary key lookups
- **QueryByIndex**: GSI queries (email, userId, projectId, etc.)
- **ScanItems**: Full table scans (needs optimization in MongoDB)
- **BatchGetItems**: Batch reads (up to 100 items)
- **BatchWriteItems**: Batch writes (up to 25 items)
- **UpdateItem**: Partial updates with expressions

---

## MongoDB Schema Design

### Database Structure

**Database Name**: `ar13_backend` (or configurable via env var)

### Collections and Indexes

#### 1. **users** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index), // Keep original ID for compatibility
  name: String,
  email: String (unique index),
  phoneNumber: String,
  role: String, // "Standard" | "Admin"
  password: String (hashed),
  designation: String?,
  createdAt: Date (indexed),
  updatedAt: Date
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ email: 1 }` - Unique
- `{ role: 1 }` - For role-based queries
- `{ createdAt: -1 }` - For sorting

#### 2. **projects** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index),
  title: String,
  description: String,
  ownerId: String (indexed),
  membersIds: [String] (indexed),
  deadLine: Date,
  logoUrl: String?,
  created: Date,
  updated: Date?
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ ownerId: 1 }` - For owner queries
- `{ membersIds: 1 }` - For member queries
- `{ created: -1 }` - For sorting

#### 3. **tasks** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index),
  subject: String,
  code: String,
  status: String (indexed),
  deadline: Date (indexed),
  priority: String,
  progress: Number? (0-100),
  assignTo: String? (indexed),
  projectId: String (indexed),
  description: String?,
  timeSpent: [{
    date: String,
    timeSpent: Number,
    userId: String,
    description: String?
  }],
  fileAttachments: [{
    fileName: String,
    originalName: String,
    fileSize: Number,
    mimeType: String,
    uploadDate: Date,
    uploadedBy: String,
    fileUrl: String
  }],
  activityLogs: [{
    id: String,
    type: String,
    timestamp: Date,
    userId: String,
    userName: String?,
    description: String,
    metadata: Object?
  }],
  created: Date,
  updated: Date?
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ projectId: 1 }` - For project queries
- `{ assignTo: 1 }` - For assignment queries
- `{ status: 1 }` - For status filtering
- `{ deadline: 1 }` - For deadline sorting
- `{ projectId: 1, status: 1 }` - Compound index

#### 4. **notifications** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index),
  userId: String (indexed),
  title: String,
  message: String,
  type: String,
  read: Boolean (indexed),
  createdAt: Date
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ userId: 1, read: 1 }` - Compound for user notifications
- `{ createdAt: -1 }` - For sorting

#### 5. **calendar_events** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index),
  title: String,
  description: String?,
  start: Date (indexed),
  end: Date (indexed),
  allDay: Boolean,
  userId: String (indexed),
  projectId: String?,
  taskId: String?,
  created: Date,
  updated: Date?
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ userId: 1, start: 1 }` - Compound for user calendar
- `{ start: 1, end: 1 }` - For date range queries

#### 6. **leaveRequests** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index),
  userId: String (indexed),
  startDate: Date,
  endDate: Date,
  type: String,
  reason: String?,
  status: String (indexed), // "pending" | "approved" | "rejected"
  approvedBy: String?,
  createdAt: Date,
  updatedAt: Date
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ userId: 1 }` - For user queries
- `{ status: 1 }` - For status filtering
- `{ userId: 1, status: 1 }` - Compound index

#### 7. **activity_logs** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index),
  entityId: String (indexed),
  entityType: String, // "task" | "project" | etc.
  type: String,
  timestamp: Date (indexed),
  userId: String,
  userName: String?,
  description: String,
  metadata: Object?
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ entityId: 1 }` - For entity queries
- `{ timestamp: -1 }` - For sorting
- `{ entityId: 1, timestamp: -1 }` - Compound index

#### 8. **info_portal** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index),
  folders: [{
    id: String,
    name: String,
    color: String,
    pages: [{
      id: String,
      title: String,
      content: String,
      folderId: String,
      createdAt: Date,
      updatedAt: Date
    }]
  }]
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ "folders.id": 1 }` - For folder lookups
- `{ "folders.pages.id": 1 }` - For page lookups

#### 9. **project_details** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index),
  projectId: String (indexed, unique),
  // Additional project detail fields
  created: Date,
  updated: Date?
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ projectId: 1 }` - Unique

#### 10. **user_account_links** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index),
  userId: String (indexed),
  provider: String, // "google" | etc.
  providerId: String,
  email: String,
  createdAt: Date
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ userId: 1 }` - For user queries
- `{ provider: 1, providerId: 1 }` - Compound unique for provider lookup

#### 11. **signupInvitations** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index),
  email: String (indexed),
  token: String (unique index),
  expiresAt: Date (indexed),
  used: Boolean,
  createdAt: Date
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ email: 1 }` - For email lookup
- `{ token: 1 }` - Unique for token validation
- `{ expiresAt: 1 }` - TTL index for cleanup

#### 12. **role_permissions** Collection
```javascript
{
  _id: ObjectId,
  id: String (unique index),
  role: String (indexed),
  permission: String,
  createdAt: Date,
  updatedAt: Date
}
```

**Indexes:**
- `{ id: 1 }` - Unique
- `{ role: 1 }` - For role queries
- `{ role: 1, permission: 1 }` - Compound unique

---

## Migration Strategy

### Approach: Dual-Write with Gradual Cutover

**Phase 1: Dual-Write Mode**
- Write to both DynamoDB and MongoDB
- Read from DynamoDB (source of truth)
- Verify data consistency

**Phase 2: Read from MongoDB**
- Switch reads to MongoDB
- Keep writing to both (safety net)
- Monitor for issues

**Phase 3: MongoDB Only**
- Write only to MongoDB
- DynamoDB becomes backup
- Monitor for 1-2 weeks

**Phase 4: Cleanup**
- Remove DynamoDB code
- Archive DynamoDB data
- Update documentation

### Data Migration Strategy

1. **Initial Data Export**: Export all data from DynamoDB
2. **Data Transformation**: Convert DynamoDB format to MongoDB format
3. **Bulk Import**: Import into MongoDB with validation
4. **Data Verification**: Compare record counts and sample data
5. **Incremental Sync**: Sync any changes during migration window

---

## Implementation Phases

### Phase 1: Infrastructure Setup (Week 1)

**Tasks:**
1. Set up MongoDB instance (local/dev/staging/prod)
2. Install MongoDB Go driver
3. Create database and collections
4. Create indexes
5. Set up connection pooling
6. Add MongoDB configuration to config

**Deliverables:**
- MongoDB instance running
- Connection code in `pkg/mongodb/`
- Configuration updated
- Indexes created

### Phase 2: Repository Abstraction (Week 2)

**Tasks:**
1. Create repository interface layer
2. Implement MongoDB base repository
3. Create MongoDB versions of all repositories
4. Implement dual-write mechanism
5. Add feature flag for database selection

**Deliverables:**
- `internal/repos/mongodb_base.go`
- MongoDB implementations for all 12 repositories
- Dual-write service layer
- Feature flag system

### Phase 3: Data Migration Scripts (Week 2-3)

**Tasks:**
1. Create data export script from DynamoDB
2. Create data transformation script
3. Create MongoDB import script
4. Create data verification script
5. Test migration on staging

**Deliverables:**
- `scripts/export_dynamodb_data.go`
- `scripts/transform_to_mongodb.go`
- `scripts/import_to_mongodb.go`
- `scripts/verify_migration.go`

### Phase 4: Dual-Write Implementation (Week 3)

**Tasks:**
1. Update all services to write to both databases
2. Add error handling and retry logic
3. Add monitoring and logging
4. Test dual-write in staging

**Deliverables:**
- All writes go to both databases
- Monitoring dashboard
- Error handling in place

### Phase 5: Read Migration (Week 4)

**Tasks:**
1. Switch reads to MongoDB (feature flag)
2. Monitor performance and errors
3. Fix any issues
4. Gradually enable for all users

**Deliverables:**
- Reads from MongoDB
- Performance metrics
- Issue tracking

### Phase 6: Cutover (Week 5)

**Tasks:**
1. Stop writing to DynamoDB
2. Final data sync
3. Monitor MongoDB-only operation
4. Archive DynamoDB data

**Deliverables:**
- MongoDB-only operation
- DynamoDB archived
- Migration complete

### Phase 7: Cleanup (Week 6)

**Tasks:**
1. Remove DynamoDB code
2. Remove feature flags
3. Update documentation
4. Performance optimization

**Deliverables:**
- Clean codebase
- Updated docs
- Optimized queries

---

## Data Migration Process

### Step 1: Export from DynamoDB

```go
// scripts/export_dynamodb_data.go
// Export all tables to JSON files
// Handle pagination for large tables
// Include metadata (table name, export date)
```

### Step 2: Transform Data

**Key Transformations:**
- Convert DynamoDB attribute values to MongoDB documents
- Handle nested structures (lists, maps)
- Convert date strings to Date objects
- Preserve original `id` field
- Add `_id` as ObjectId

### Step 3: Import to MongoDB

```go
// scripts/import_to_mongodb.go
// Bulk insert with batching
// Handle duplicates
// Create indexes after import
// Validate data integrity
```

### Step 4: Verify Migration

- Record count comparison
- Sample data validation
- Index verification
- Query result comparison

---

## Code Changes Required

### 1. Add MongoDB Driver

```bash
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/bson
```

### 2. Create MongoDB Package

**File**: `pkg/mongodb/mongodb.go`
- Connection management
- Database access
- Health checks

### 3. Update Configuration

**File**: `internal/config/config.go`
```go
type Config struct {
    // ... existing fields
    
    // MongoDB
    MongoDBURI      string
    MongoDBDatabase string
}
```

### 4. Create MongoDB Base Repository

**File**: `internal/repos/mongodb_base.go`
- Similar interface to `DynamoBaseRepo`
- MongoDB-specific implementations
- Index management

### 5. Create MongoDB Repositories

For each repository in `internal/repos/`:
- Create MongoDB version: `*_repo_mongodb.go`
- Implement same interface
- Handle MongoDB-specific operations

### 6. Update Service Layer

**Option A: Interface-Based**
- Create repository interfaces
- Inject repository implementation
- Switch via dependency injection

**Option B: Feature Flag**
- Add feature flag check
- Route to appropriate repository
- Gradual migration

### 7. Update Models

- Add MongoDB tags (`bson`)
- Keep DynamoDB tags for backward compatibility
- Handle date/time conversions

### 8. Update Main Application

**File**: `cmd/server/main.go`
- Initialize MongoDB connection
- Initialize repositories based on config
- Add health check endpoints

---

## Testing Strategy

### Unit Tests
- Test each MongoDB repository method
- Test data transformations
- Test error handling

### Integration Tests
- Test with real MongoDB instance
- Test dual-write scenarios
- Test migration scripts

### End-to-End Tests
- Test complete user flows
- Test data consistency
- Test performance

### Load Tests
- Compare MongoDB vs DynamoDB performance
- Test concurrent operations
- Test batch operations

### Data Validation Tests
- Verify all data migrated correctly
- Test edge cases (nulls, empty arrays, etc.)
- Test date/time handling

---

## Rollback Plan

### Rollback Triggers
- Data loss detected
- Performance degradation > 50%
- Critical errors > 1% of requests
- User complaints

### Rollback Steps
1. **Immediate**: Switch reads back to DynamoDB (feature flag)
2. **Short-term**: Stop writing to MongoDB, continue with DynamoDB
3. **Data Recovery**: Restore from DynamoDB if needed
4. **Investigation**: Analyze issues, fix MongoDB setup
5. **Retry**: Attempt migration again after fixes

### Rollback Testing
- Test rollback procedure in staging
- Document rollback steps
- Create rollback scripts

---

## Timeline and Dependencies

### Estimated Timeline: 6-8 Weeks

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Infrastructure Setup | 1 week | MongoDB instance access |
| Repository Abstraction | 1 week | MongoDB driver installed |
| Data Migration Scripts | 1-2 weeks | Repository abstraction complete |
| Dual-Write Implementation | 1 week | All repositories ready |
| Read Migration | 1 week | Data migration complete |
| Cutover | 1 week | Read migration stable |
| Cleanup | 1 week | Cutover successful |

### Dependencies
- MongoDB instance (local/dev/staging/prod)
- MongoDB Go driver
- Access to DynamoDB for export
- Staging environment for testing
- Backup/restore procedures

---

## Risk Assessment

### High Risk
- **Data Loss**: Mitigated by dual-write and backups
- **Performance Issues**: Mitigated by load testing
- **Downtime**: Mitigated by gradual migration

### Medium Risk
- **Data Inconsistency**: Mitigated by verification scripts
- **Migration Time**: Mitigated by batch processing
- **Code Complexity**: Mitigated by clean abstractions

### Low Risk
- **Index Creation**: Can be done incrementally
- **Configuration Changes**: Well-documented
- **Team Training**: MongoDB is well-documented

---

## Success Criteria

### Technical
- ✅ All data migrated successfully
- ✅ All queries working correctly
- ✅ Performance equal or better than DynamoDB
- ✅ Zero data loss
- ✅ All tests passing

### Business
- ✅ Zero downtime during migration
- ✅ No user-visible errors
- ✅ Cost reduction achieved
- ✅ Improved developer experience

---

## Next Steps

1. **Review and Approve Plan**: Get stakeholder approval
2. **Set Up MongoDB**: Provision MongoDB instance
3. **Create Proof of Concept**: Migrate one table as POC
4. **Begin Phase 1**: Start infrastructure setup
5. **Regular Updates**: Weekly progress reports

---

## Appendix

### A. MongoDB Connection String Format
```
mongodb://[username:password@]host1[:port1][,...hostN[:portN]][/[defaultauthdb][?options]]
```

### B. Sample MongoDB Repository Code
See `docs/migration/MONGODB_REPOSITORY_EXAMPLES.md` (to be created)

### C. Migration Checklist
See `docs/migration/MIGRATION_CHECKLIST.md` (to be created)

### D. Troubleshooting Guide
See `docs/migration/MIGRATION_TROUBLESHOOTING.md` (to be created)

---

**Document Version**: 1.0  
**Last Updated**: 2025-01-XX  
**Author**: Migration Planning Team  
**Status**: Draft - Pending Review

