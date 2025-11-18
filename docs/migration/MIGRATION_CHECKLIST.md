# DynamoDB to MongoDB Migration Checklist

This checklist tracks progress through the migration phases.

## Pre-Migration

### Planning
- [ ] Review and approve migration plan
- [ ] Identify stakeholders and get buy-in
- [ ] Set up project tracking (Jira/GitHub Issues)
- [ ] Schedule migration timeline
- [ ] Allocate resources (developers, DevOps)

### Infrastructure
- [ ] Provision MongoDB instance (dev)
- [ ] Provision MongoDB instance (staging)
- [ ] Provision MongoDB instance (production)
- [ ] Set up MongoDB monitoring
- [ ] Configure MongoDB backups
- [ ] Set up MongoDB connection pooling
- [ ] Test MongoDB connectivity from application

### Environment Setup
- [ ] Add MongoDB URI to dev environment variables
- [ ] Add MongoDB URI to staging environment variables
- [ ] Add MongoDB URI to production environment variables
- [ ] Add MongoDB database name to config
- [ ] Update `.env.example` with MongoDB variables

---

## Phase 1: Infrastructure Setup

### MongoDB Driver
- [ ] Install MongoDB Go driver (`go.mongodb.org/mongo-driver`)
- [ ] Update `go.mod` and `go.sum`
- [ ] Test driver installation

### MongoDB Package
- [ ] Create `pkg/mongodb/mongodb.go`
- [ ] Implement connection initialization
- [ ] Implement connection health check
- [ ] Add connection pooling configuration
- [ ] Add error handling
- [ ] Write unit tests for connection

### Configuration
- [ ] Add `MongoDBURI` to `internal/config/config.go`
- [ ] Add `MongoDBDatabase` to `internal/config/config.go`
- [ ] Update `.env` files with MongoDB config
- [ ] Test configuration loading

### Database Setup
- [ ] Create database in MongoDB
- [ ] Create all 12 collections
- [ ] Create all indexes (see migration plan)
- [ ] Verify index creation
- [ ] Document index strategy

---

## Phase 2: Repository Abstraction

### Base Repository
- [ ] Create `internal/repos/mongodb_base.go`
- [ ] Implement `GetByID` method
- [ ] Implement `GetAll` method
- [ ] Implement `Add` method
- [ ] Implement `Update` method
- [ ] Implement `Delete` method
- [ ] Implement `Query` method (replaces GSI queries)
- [ ] Implement `BatchGet` method
- [ ] Implement `BatchWrite` method
- [ ] Write unit tests for base repository

### Repository Interfaces
- [ ] Create repository interfaces (optional but recommended)
- [ ] Define common interface methods
- [ ] Document interface contracts

### MongoDB Repositories
- [ ] Create `internal/repos/user_repo_mongodb.go`
- [ ] Create `internal/repos/project_repo_mongodb.go`
- [ ] Create `internal/repos/task_repo_mongodb.go`
- [ ] Create `internal/repos/notification_repo_mongodb.go`
- [ ] Create `internal/repos/calendar_repo_mongodb.go`
- [ ] Create `internal/repos/vacation_repo_mongodb.go`
- [ ] Create `internal/repos/activity_log_repo_mongodb.go`
- [ ] Create `internal/repos/info_portal_repo_mongodb.go`
- [ ] Create `internal/repos/project_details_repo_mongodb.go`
- [ ] Create `internal/repos/user_account_link_repo_mongodb.go`
- [ ] Create `internal/repos/signup_invitation_repo_mongodb.go`
- [ ] Create `internal/repos/role_permission_repo_mongodb.go`

### Repository Implementation
For each repository:
- [ ] Implement all methods from DynamoDB version
- [ ] Handle MongoDB-specific operations
- [ ] Add proper error handling
- [ ] Write unit tests
- [ ] Test with real MongoDB instance

### Feature Flag System
- [ ] Create feature flag configuration
- [ ] Add `USE_MONGODB` environment variable
- [ ] Implement repository factory/selector
- [ ] Test feature flag switching

---

## Phase 3: Data Migration Scripts

### Export Script
- [ ] Create `scripts/export_dynamodb_data.go`
- [ ] Export `users` table
- [ ] Export `projects` table
- [ ] Export `tasks` table
- [ ] Export `notifications` table
- [ ] Export `calendar_events` table
- [ ] Export `leaveRequests` table
- [ ] Export `activity_logs` table
- [ ] Export `info-portal` table
- [ ] Export `project_details` table
- [ ] Export `user_account_links` table
- [ ] Export `signupInvitations` table
- [ ] Export `role_permissions` table
- [ ] Handle pagination for large tables
- [ ] Add progress logging
- [ ] Test export script

### Transform Script
- [ ] Create `scripts/transform_to_mongodb.go`
- [ ] Transform DynamoDB attribute values
- [ ] Convert date strings to Date objects
- [ ] Handle nested structures
- [ ] Preserve original IDs
- [ ] Add `_id` as ObjectId
- [ ] Validate transformed data
- [ ] Test transformation script

### Import Script
- [ ] Create `scripts/import_to_mongodb.go`
- [ ] Bulk insert with batching
- [ ] Handle duplicate IDs
- [ ] Add progress logging
- [ ] Validate import success
- [ ] Test import script

### Verification Script
- [ ] Create `scripts/verify_migration.go`
- [ ] Compare record counts
- [ ] Sample data validation
- [ ] Index verification
- [ ] Query result comparison
- [ ] Generate verification report
- [ ] Test verification script

### Staging Migration
- [ ] Run export on staging DynamoDB
- [ ] Transform data
- [ ] Import to staging MongoDB
- [ ] Verify migration
- [ ] Fix any issues
- [ ] Document issues and solutions

---

## Phase 4: Dual-Write Implementation

### Service Layer Updates
- [ ] Update `UserService` for dual-write
- [ ] Update `ProjectService` for dual-write
- [ ] Update `TaskService` for dual-write
- [ ] Update `NotificationService` for dual-write
- [ ] Update `CalendarService` for dual-write
- [ ] Update `VacationService` for dual-write
- [ ] Update `ActivityLogService` for dual-write
- [ ] Update `InfoPortalService` for dual-write
- [ ] Update `ProjectDetailsService` for dual-write
- [ ] Update `UserAccountLinkService` for dual-write
- [ ] Update `SignupInvitationService` for dual-write
- [ ] Update `PermissionService` for dual-write

### Error Handling
- [ ] Add retry logic for MongoDB writes
- [ ] Add error logging
- [ ] Add metrics/monitoring
- [ ] Handle partial failures
- [ ] Alert on write failures

### Monitoring
- [ ] Set up MongoDB write metrics
- [ ] Set up DynamoDB write metrics
- [ ] Compare write performance
- [ ] Monitor error rates
- [ ] Create monitoring dashboard

### Testing
- [ ] Test dual-write in dev
- [ ] Test dual-write in staging
- [ ] Load test dual-write
- [ ] Test error scenarios
- [ ] Verify data consistency

---

## Phase 5: Read Migration

### Read Switching
- [ ] Add feature flag for MongoDB reads
- [ ] Switch `UserService` reads to MongoDB
- [ ] Switch `ProjectService` reads to MongoDB
- [ ] Switch `TaskService` reads to MongoDB
- [ ] Switch `NotificationService` reads to MongoDB
- [ ] Switch `CalendarService` reads to MongoDB
- [ ] Switch `VacationService` reads to MongoDB
- [ ] Switch `ActivityLogService` reads to MongoDB
- [ ] Switch `InfoPortalService` reads to MongoDB
- [ ] Switch `ProjectDetailsService` reads to MongoDB
- [ ] Switch `UserAccountLinkService` reads to MongoDB
- [ ] Switch `SignupInvitationService` reads to MongoDB
- [ ] Switch `PermissionService` reads to MongoDB

### Gradual Rollout
- [ ] Enable MongoDB reads for 10% of requests
- [ ] Monitor for 24 hours
- [ ] Enable MongoDB reads for 50% of requests
- [ ] Monitor for 24 hours
- [ ] Enable MongoDB reads for 100% of requests
- [ ] Monitor for 48 hours

### Performance Monitoring
- [ ] Compare query performance
- [ ] Monitor MongoDB query times
- [ ] Monitor error rates
- [ ] Optimize slow queries
- [ ] Add missing indexes if needed

---

## Phase 6: Cutover

### Production Data Migration
- [ ] Schedule maintenance window (if needed)
- [ ] Export production DynamoDB data
- [ ] Transform production data
- [ ] Import to production MongoDB
- [ ] Verify production migration
- [ ] Run verification script

### Final Sync
- [ ] Stop new writes temporarily
- [ ] Sync any remaining changes
- [ ] Verify data consistency
- [ ] Resume writes

### DynamoDB Write Removal
- [ ] Remove DynamoDB writes from `UserService`
- [ ] Remove DynamoDB writes from `ProjectService`
- [ ] Remove DynamoDB writes from `TaskService`
- [ ] Remove DynamoDB writes from `NotificationService`
- [ ] Remove DynamoDB writes from `CalendarService`
- [ ] Remove DynamoDB writes from `VacationService`
- [ ] Remove DynamoDB writes from `ActivityLogService`
- [ ] Remove DynamoDB writes from `InfoPortalService`
- [ ] Remove DynamoDB writes from `ProjectDetailsService`
- [ ] Remove DynamoDB writes from `UserAccountLinkService`
- [ ] Remove DynamoDB writes from `SignupInvitationService`
- [ ] Remove DynamoDB writes from `PermissionService`

### Monitoring
- [ ] Monitor MongoDB-only operation for 24 hours
- [ ] Monitor MongoDB-only operation for 1 week
- [ ] Monitor error rates
- [ ] Monitor performance
- [ ] Address any issues

### Backup
- [ ] Export final DynamoDB data
- [ ] Store DynamoDB backup securely
- [ ] Document backup location
- [ ] Set retention policy

---

## Phase 7: Cleanup

### Code Removal
- [ ] Remove `pkg/dynamodb/` package
- [ ] Remove `internal/repos/dynamodb_base.go`
- [ ] Remove all `*_repo.go` files (DynamoDB versions)
- [ ] Remove DynamoDB imports
- [ ] Remove DynamoDB configuration
- [ ] Remove feature flags
- [ ] Remove dual-write code

### Dependency Cleanup
- [ ] Remove AWS SDK DynamoDB dependencies from `go.mod`
- [ ] Run `go mod tidy`
- [ ] Verify no DynamoDB references remain

### Documentation
- [ ] Update README.md
- [ ] Update deployment guides
- [ ] Update API documentation
- [ ] Update architecture diagrams
- [ ] Archive old DynamoDB documentation

### Testing
- [ ] Run full test suite
- [ ] Fix any broken tests
- [ ] Update test fixtures
- [ ] Update integration tests

### Performance Optimization
- [ ] Analyze slow queries
- [ ] Add missing indexes
- [ ] Optimize aggregation pipelines
- [ ] Review connection pooling
- [ ] Optimize batch operations

---

## Post-Migration

### Validation
- [ ] All features working correctly
- [ ] Performance meets or exceeds DynamoDB
- [ ] No data loss
- [ ] All tests passing
- [ ] Documentation updated

### Cost Analysis
- [ ] Compare MongoDB costs vs DynamoDB
- [ ] Document cost savings
- [ ] Update budget projections

### Team Training
- [ ] MongoDB training session
- [ ] Update development guidelines
- [ ] Share migration learnings

### Retrospective
- [ ] Conduct migration retrospective
- [ ] Document lessons learned
- [ ] Update migration plan for future reference

---

## Rollback Checklist (If Needed)

### Immediate Rollback
- [ ] Switch reads back to DynamoDB
- [ ] Stop MongoDB writes
- [ ] Verify DynamoDB is working
- [ ] Notify team

### Data Recovery
- [ ] Assess data loss (if any)
- [ ] Restore from DynamoDB backup if needed
- [ ] Verify data integrity

### Investigation
- [ ] Analyze root cause
- [ ] Document issues
- [ ] Create fix plan
- [ ] Test fixes in staging

---

**Last Updated**: 2025-01-XX  
**Status**: In Progress

