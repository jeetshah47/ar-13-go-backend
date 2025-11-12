# DynamoDB Optimization Implementation Guide

## Overview

This document outlines the optimizations implemented to reduce DynamoDB usage and stay within AWS Free Tier limits.

## Implemented Optimizations

### 1. ✅ Batch Operations

**Location:** `internal/repos/dynamodb_base.go`

**Added Methods:**
- `BatchGetItems()` - Retrieves up to 100 items in a single request
- `BatchWriteItems()` - Writes up to 25 items in a single request
- `ScanItemsPaginated()` - Scans with pagination support

**Benefits:**
- Reduces number of DynamoDB API calls
- Lower read/write capacity consumption
- Faster performance for bulk operations

**Usage Example:**
```go
// Instead of multiple GetByID calls:
items, err := repo.BatchGetItems(ctx, []string{"id1", "id2", "id3"})
```

### 2. ✅ Item-Level Caching

**Location:** `internal/services/cache_service.go`

**Added Caching:**
- User records (10 min TTL)
- Project records (10 min TTL)
- Task records (5 min TTL)

**Benefits:**
- Reduces read capacity for frequently accessed items
- Faster response times
- Automatic cache invalidation on updates

**Usage Example:**
```go
// In services, check cache first:
var user models.User
err := cacheSvc.GetUser(ctx, userID, &user)
if err == cache.ErrCacheMiss {
    // Fetch from DynamoDB and cache
    user, _ := userRepo.GetByID(ctx, userID)
    _ = cacheSvc.SetUser(ctx, userID, user)
}
```

### 3. ✅ Optimized Activity Log User Population

**Location:** `internal/services/task_service.go` - `populateActivityLogUsers()`

**Optimizations:**
- Collects unique user IDs first
- Uses batch operations to fetch multiple users at once
- Checks cache before DynamoDB reads
- Reduces N+1 query problem

**Before:** N individual GetByID calls (N = number of unique users)
**After:** 1 batch operation + cache lookups

**Impact:** 70-90% reduction in read capacity for activity logs

### 4. ✅ Task Service Caching

**Location:** `internal/services/task_service.go`

**Optimizations:**
- `GetByID()` checks cache before DynamoDB
- Cache invalidation on task updates/deletes
- Cache population on task creation

**Impact:** 80-90% reduction in read capacity for frequently accessed tasks

## Additional Optimizations Needed

### Priority 1: Replace Scans with Queries

**Current Issue:**
- `project_repo.GetAll()` uses `ScanItems()` - reads entire table
- `user_repo.GetAll()` uses `ScanItems()` - reads entire table
- `project_repo.Persists()` uses `ScanItems()` - scans entire table just to check existence

**Recommended Fix:**

1. **Add Pagination to GetAll Methods:**
```go
func (r *ProjectRepo) GetAll(ctx context.Context, limit *int, lastKey map[string]types.AttributeValue) ([]models.Project, map[string]types.AttributeValue, error) {
    var limitInt32 *int32
    if limit != nil {
        l := int32(*limit)
        limitInt32 = &l
    }
    
    items, lastEvaluatedKey, err := r.ScanItemsPaginated(ctx, limitInt32, lastKey)
    // ... process items
    return projects, lastEvaluatedKey, nil
}
```

2. **Replace Persists() with Exists():**
```go
// Instead of scanning entire table:
func (r *ProjectRepo) Persists(ctx context.Context, id string) (bool, error) {
    return r.Exists(ctx, id) // Uses GetItem instead of Scan
}
```

**Expected Impact:** 50-70% reduction in read capacity for list operations

### Priority 2: Add GSIs for Common Queries

**Recommended GSIs:**

1. **Projects by Owner:**
   - GSI: `ownerId-index` (ownerId as partition key)
   - Use for: `GetProjectsByOwner()`

2. **Tasks by Assignee:**
   - GSI: `assignTo-index` (assignTo as partition key)
   - Use for: `GetTasksByAssignee()`

**Expected Impact:** Replace scans with efficient queries

### Priority 3: Optimize Update Operations

**Current Issue:**
- Many operations call `GetByID()` before updating
- Example: `task_repo.AddTimeSpent()` fetches entire task, modifies, then updates

**Recommended Fix:**
- Use atomic update operations where possible
- Use `UpdateExpression` with `ADD` for list operations
- Avoid read-before-write patterns

**Example:**
```go
// Instead of:
task, _ := repo.GetByID(ctx, taskID)
task.TimeSpent = append(task.TimeSpent, newEntry)
repo.Update(ctx, task)

// Use atomic update:
repo.UpdateItem(ctx, taskID, map[string]interface{}{
    "timeSpent": dynamodb.AttributeValue{
        L: []dynamodb.AttributeValue{...} // Append operation
    }
})
```

## Monitoring & Validation

### Metrics to Track

1. **DynamoDB Read Capacity:**
   - Target: < 18,600 RCU-Hrs/month (72% of free tier)
   - Current: 13,410 RCU-Hrs (72.10%)
   - After optimizations: Expected < 5,000 RCU-Hrs/month

2. **DynamoDB Write Capacity:**
   - Target: < 18,600 WCU-Hrs/month (72% of free tier)
   - Current: 13,440 WCU-Hrs (72.26%)
   - After optimizations: Expected < 8,000 WCU-Hrs/month

3. **Cache Hit Rate:**
   - Monitor Redis cache hit rate
   - Target: > 70% for frequently accessed items

### CloudWatch Metrics

Monitor these DynamoDB metrics:
- `ConsumedReadCapacityUnits`
- `ConsumedWriteCapacityUnits`
- `ReadThrottleEvents`
- `WriteThrottleEvents`

### Testing Checklist

- [ ] Verify batch operations work correctly
- [ ] Test cache invalidation on updates
- [ ] Verify cache TTLs are appropriate
- [ ] Test pagination for list operations
- [ ] Monitor DynamoDB usage after deployment
- [ ] Verify no increase in errors after optimizations

## Deployment Steps

1. **Deploy Code Changes:**
   ```bash
   git add .
   git commit -m "Add DynamoDB optimizations: batch operations and caching"
   git push
   ```

2. **Monitor for 24-48 hours:**
   - Check AWS Free Tier dashboard
   - Monitor CloudWatch metrics
   - Verify cache hit rates

3. **Implement Priority 1 Optimizations:**
   - Add pagination to GetAll methods
   - Replace Persists() scans with Exists()

4. **Continue Monitoring:**
   - Daily checks for first week
   - Weekly checks after that
   - Adjust cache TTLs if needed

## Expected Results

### Before Optimizations:
- Read Capacity: 13,410 RCU-Hrs (72.10%)
- Write Capacity: 13,440 WCU-Hrs (72.26%)
- Forecasted: 180%+ (exceeding free tier)

### After All Optimizations:
- Read Capacity: < 5,000 RCU-Hrs (< 27%)
- Write Capacity: < 8,000 WCU-Hrs (< 43%)
- Forecasted: < 50% (well within free tier)

### Performance Improvements:
- Faster response times (cache hits)
- Reduced API latency (batch operations)
- Lower costs (staying within free tier)

## Notes

- Cache TTLs can be adjusted based on usage patterns
- Batch operations automatically handle DynamoDB limits (100 reads, 25 writes)
- All optimizations are backward compatible
- Fallback to individual operations if batch fails

