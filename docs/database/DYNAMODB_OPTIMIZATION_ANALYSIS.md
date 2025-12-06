# DynamoDB Usage Optimization Analysis

## Current Usage Status

**Free Tier Limits:**
- Write Capacity: 18,600 WCU-Hrs/month (forever free)
- Read Capacity: 18,600 RCU-Hrs/month (forever free)

**Current Usage (Month-to-Date):**
- Write Capacity: 13,440 WCU-Hrs (72.26% used)
- Read Capacity: 13,410 RCU-Hrs (72.10% used)

**Forecasted Usage:**
- Write Capacity: 33,600 WCU-Hrs (180.65% - **EXCEEDING FREE TIER**)
- Read Capacity: 33,525 RCU-Hrs (180.24% - **EXCEEDING FREE TIER**)

## Critical Issues Identified

### 1. ⚠️ Excessive Use of Scan Operations

**Problem:** Scan operations read through entire tables, consuming capacity units for every item scanned.

**Current Usage:**
- `project_repo.go`: `GetAll()` uses `ScanItems()` - scans entire projects table
- `project_repo.go`: `Persists()` uses `ScanItems()` - scans entire table just to check existence
- `user_repo.go`: `GetAll()` uses `ScanItems()` - scans entire users table
- `vacation_repo.go`: Multiple `ScanItems()` calls
- `info_portal_repo.go`: Uses `ScanItems()`
- `user_account_link_repo.go`: Uses `ScanItems()`

**Impact:**
- Each scan reads ALL items in the table
- If you have 100 projects, scanning reads all 100 items
- If you have 50 users, scanning reads all 50 items
- **This is the #1 cause of high read capacity usage**

**Solution:**
- Replace scans with queries using GSIs where possible
- Add pagination to limit scan results
- Use `ProjectionExpression` to only fetch needed attributes
- Consider adding GSIs for common query patterns

### 2. ⚠️ No Batch Operations

**Problem:** Every read/write is an individual operation, causing multiple round trips.

**Current Pattern:**
```go
// In task_service.go - Multiple individual calls
task, err := r.taskRepo.GetByID(ctx, projectID, taskID)  // 1 read
user, err := s.userRepo.GetByID(ctx, log.UserID)          // 1 read
project, err := projectRepo.GetByID(ctx, projectID)       // 1 read
```

**Impact:**
- 3 separate DynamoDB calls = 3 read capacity units
- Network latency multiplied by number of calls
- Higher costs and slower performance

**Solution:**
- Implement `BatchGetItem` for multiple reads
- Implement `BatchWriteItem` for multiple writes
- Batch user lookups in activity log population

### 3. ⚠️ Redundant GetByID Calls

**Problem:** Same items are fetched multiple times in a single request.

**Examples:**
- `task_service.go`: `Update()` calls `GetByID()` to check existence, then updates
- `task_service.go`: `populateActivityLogUsers()` calls `GetByID()` for each user (no batching)
- `task_repo.go`: `AddTimeSpent()` calls `GetByID()` before updating
- `task_repo.go`: `UpdateTimeSpent()` calls `GetByID()` before updating
- `task_repo.go`: `AddFileAttachment()` calls `GetByID()` before updating

**Impact:**
- Same task fetched multiple times in one operation
- Same user fetched multiple times for different activity logs
- Wasted read capacity units

**Solution:**
- Cache items in memory during single request
- Use batch operations for multiple items
- Only fetch when absolutely necessary

### 4. ⚠️ Limited Caching Strategy

**Current Caching:**
- ✅ Dashboard stats (3 min TTL)
- ✅ Calendar events (10 min TTL)
- ✅ Project stats (2 min TTL)
- ✅ Activity logs (1 min TTL)

**Missing Caching:**
- ❌ Individual user records
- ❌ Individual project records
- ❌ Individual task records
- ❌ Frequently accessed lookups

**Impact:**
- Every API call fetches from DynamoDB
- No reduction in read capacity for hot data
- Higher costs for frequently accessed items

**Solution:**
- Add caching for user records (5-10 min TTL)
- Add caching for project records (5-10 min TTL)
- Add caching for task records (2-5 min TTL)
- Invalidate cache on updates

### 5. ⚠️ Inefficient Query Patterns

**Problem:** Some operations could use queries instead of scans.

**Examples:**
- `project_repo.Persists()`: Scans entire table to check if title exists
- `user_repo.GetAll()`: Scans entire table instead of using pagination
- `vacation_repo`: Multiple full table scans

**Solution:**
- Add GSI for project titles if needed
- Implement pagination for list operations
- Use `Limit` parameter more effectively

## Optimization Recommendations

### Priority 1: Critical (Immediate Impact)

1. **Replace Scans with Queries**
   - Add GSIs for common query patterns
   - Use `QueryByIndex` instead of `ScanItems`
   - Add pagination to all list operations

2. **Implement Batch Operations**
   - Add `BatchGetItem` for multiple reads
   - Add `BatchWriteItem` for multiple writes
   - Batch user lookups in activity logs

3. **Add Item-Level Caching**
   - Cache users, projects, tasks
   - Use Redis with appropriate TTLs
   - Invalidate on updates

### Priority 2: High Impact

4. **Optimize Redundant Calls**
   - Cache items in request context
   - Reduce `GetByID` calls in task operations
   - Batch user fetches in activity logs

5. **Add Pagination**
   - Limit scan results
   - Use `Limit` parameter
   - Implement cursor-based pagination

### Priority 3: Medium Impact

6. **Use ProjectionExpression**
   - Only fetch needed attributes
   - Reduce data transfer
   - Lower read capacity usage

7. **Optimize Update Operations**
   - Use conditional updates where possible
   - Avoid read-before-write patterns
   - Use atomic operations

## Expected Impact

### After Priority 1 Optimizations:
- **Read Capacity Reduction**: 50-70% reduction
- **Write Capacity Reduction**: 20-30% reduction
- **Forecasted Usage**: Should drop below 100% of free tier

### After Priority 2 Optimizations:
- **Read Capacity Reduction**: Additional 20-30% reduction
- **Total Reduction**: 70-90% reduction in read capacity
- **Forecasted Usage**: Well within free tier limits

## Implementation Plan

See `DYNAMODB_OPTIMIZATION_IMPLEMENTATION.md` for detailed implementation steps.

## Monitoring

After implementing optimizations:
1. Monitor AWS Free Tier dashboard daily
2. Check DynamoDB CloudWatch metrics
3. Track read/write capacity usage
4. Verify forecasted usage drops below 100%

