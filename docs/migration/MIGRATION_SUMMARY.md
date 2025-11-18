# DynamoDB to MongoDB Migration - Quick Summary

## Overview

This is a quick reference guide for the DynamoDB to MongoDB migration. For detailed information, see the full migration plan.

## Key Documents

1. **[DYNAMODB_TO_MONGODB_MIGRATION_PLAN.md](./DYNAMODB_TO_MONGODB_MIGRATION_PLAN.md)** - Complete migration plan
2. **[MIGRATION_CHECKLIST.md](./MIGRATION_CHECKLIST.md)** - Step-by-step checklist
3. **[MONGODB_REPOSITORY_EXAMPLES.md](./MONGODB_REPOSITORY_EXAMPLES.md)** - Code examples
4. **[DATA_TRANSFORMATION_GUIDE.md](./DATA_TRANSFORMATION_GUIDE.md)** - Data transformation guide

## Migration Strategy

**Approach**: Dual-write with gradual cutover

1. **Phase 1**: Write to both DynamoDB and MongoDB
2. **Phase 2**: Read from MongoDB, write to both
3. **Phase 3**: Write only to MongoDB
4. **Phase 4**: Remove DynamoDB code

## Timeline

**Estimated Duration**: 6-8 weeks

| Phase | Duration | Key Activities |
|-------|----------|----------------|
| Infrastructure Setup | 1 week | MongoDB setup, driver installation |
| Repository Abstraction | 1 week | Create MongoDB repositories |
| Data Migration Scripts | 1-2 weeks | Export, transform, import scripts |
| Dual-Write | 1 week | Write to both databases |
| Read Migration | 1 week | Switch reads to MongoDB |
| Cutover | 1 week | MongoDB-only operation |
| Cleanup | 1 week | Remove DynamoDB code |

## Current State

### DynamoDB Tables (12 total)
- users
- projects
- tasks
- notifications
- calendar_events
- leaveRequests
- activity_logs
- info-portal
- project_details
- user_account_links
- signupInvitations
- role_permissions

### Architecture
- Repository pattern (`internal/repos/`)
- Service layer (`internal/services/`)
- Models with DynamoDB tags (`internal/models/`)

## MongoDB Schema

### Database
- **Name**: `ar13_backend` (configurable)

### Collections
- Same names as DynamoDB tables
- `_id` as ObjectId (auto-generated)
- Original `id` field preserved (unique index)
- Indexes for all query patterns

## Key Code Changes

### 1. Add MongoDB Driver
```bash
go get go.mongodb.org/mongo-driver/mongo
```

### 2. Create MongoDB Package
- `pkg/mongodb/mongodb.go` - Connection management

### 3. Create MongoDB Repositories
- `internal/repos/mongodb_base.go` - Base repository
- `internal/repos/*_repo_mongodb.go` - Collection-specific repos

### 4. Update Configuration
```go
type Config struct {
    MongoDBURI      string
    MongoDBDatabase string
}
```

### 5. Update Services
- Add dual-write logic
- Feature flag for database selection
- Error handling and monitoring

## Data Migration Steps

1. **Export** from DynamoDB to JSON
2. **Transform** DynamoDB format to MongoDB format
3. **Import** to MongoDB with batching
4. **Verify** record counts and sample data

## Success Criteria

- ✅ All data migrated successfully
- ✅ All queries working correctly
- ✅ Performance equal or better
- ✅ Zero data loss
- ✅ Zero downtime

## Quick Start

### 1. Set Up MongoDB
```bash
# Install MongoDB locally or use cloud service
# Update .env with connection string
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=ar13_backend
```

### 2. Create Collections and Indexes
```bash
go run scripts/create_mongodb_indexes.go
```

### 3. Export DynamoDB Data
```bash
go run scripts/export_dynamodb_data.go
```

### 4. Transform and Import
```bash
go run scripts/transform_to_mongodb.go
go run scripts/import_to_mongodb.go
```

### 5. Verify Migration
```bash
go run scripts/verify_migration.go
```

## Rollback Plan

If issues occur:
1. Switch reads back to DynamoDB (feature flag)
2. Stop MongoDB writes
3. Investigate and fix issues
4. Retry migration

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Data Loss | Dual-write, backups, verification |
| Performance Issues | Load testing, optimization |
| Downtime | Gradual migration, feature flags |

## Next Steps

1. Review and approve migration plan
2. Set up MongoDB instance
3. Create proof of concept (one table)
4. Begin Phase 1 implementation

## Support

For questions or issues:
- Review detailed migration plan
- Check code examples
- Consult MongoDB documentation
- Review transformation guide

---

**Status**: Planning Phase  
**Last Updated**: 2025-01-XX
