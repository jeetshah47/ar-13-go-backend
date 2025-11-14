# MongoDB Index Optimization Guide

## Overview

This document explains the index optimization strategy used to minimize write overhead while maintaining query performance.

## Index Overhead

### Write Performance Impact

Every index adds overhead to write operations:
- **Insert**: Must write to collection + update all indexes
- **Update**: Must update collection + all indexes that include modified fields
- **Delete**: Must remove from collection + remove from all indexes

**Rule of Thumb**: Each additional index can add 5-10% write overhead.

### Storage Impact

- Indexes consume disk space (typically 10-20% of collection size)
- Indexes are kept in memory for performance
- More indexes = more memory usage

## Optimization Strategy

### ✅ Indexes Created (Essential Only)

#### 1. **users** Collection
- `id` (unique) - Primary key lookups
- `email` (unique) - Used in `GetByEmail()` query

**Removed**: `role`, `createdAt` - Not queried directly, only used in scans

#### 2. **projects** Collection
- `id` (unique) - Primary key lookups

**Removed**: `ownerId`, `membersIds`, `created` - Only used in scans, not indexed queries

#### 3. **tasks** Collection
- `id` (unique) - Primary key lookups
- `projectId` - Used in `GetAll()` by project

**Removed**: `assignTo`, `status`, `deadline`, compound indexes - Not queried via indexes

#### 4. **notifications** Collection
- `id` (unique) - Primary key lookups
- `userId` - Used in multiple queries
- `userId + read` (compound) - Optimizes `GetUnread()` queries

**Removed**: `createdAt` - Sorting can be done without index for small result sets

#### 5. **calendar_events** Collection
- `id` (unique) - Primary key lookups

**Removed**: `userId+start`, `start+end` - `GetByMonth()` uses scan with filter, not indexed query

#### 6. **leaveRequests** Collection
- `id` (unique) - Primary key lookups
- `userId` - Used in `GetByUserID()`
- `status` - Used in `GetByStatus()`

**Removed**: Compound `userId+status` - Not used together in queries

#### 7. **activity_logs** Collection
- `id` (unique) - Primary key lookups
- `entityId` - Used in `GetByEntity()`
- `entityId + timestamp` (compound) - Optimizes sorted results by entity

**Removed**: Standalone `timestamp` - Not queried alone

#### 8. **project_details** Collection
- `id` (unique) - Primary key lookups
- `projectId` (unique) - Used in queries

#### 9. **user_account_links** Collection
- `id` (unique) - Primary key lookups
- `userId` - Used in `GetByUserID()`
- `provider + providerId` (compound unique) - Used in provider lookups

#### 10. **signupInvitations** Collection
- `id` (unique) - Primary key lookups
- `email` - Used in `GetByEmail()`
- `token` (unique) - Used in `GetByToken()`
- `expiresAt` (TTL) - Automatic cleanup of expired invitations

#### 11. **role_permissions** Collection
- `id` (unique) - Primary key lookups
- `role` - Used in `GetByRole()`
- `role + permission` (compound unique) - Prevents duplicates

#### 12. **info-portal** Collection
- `id` (unique) - Primary key lookups

**Removed**: Nested `folders.id`, `folders.pages.id` - Nested array indexes add significant write overhead

## Index Count Summary

| Collection | Indexes Created | Removed | Reason |
|------------|----------------|---------|--------|
| users | 2 | 2 | Removed role, createdAt (not queried) |
| projects | 1 | 3 | Removed ownerId, membersIds, created (scans only) |
| tasks | 2 | 4 | Removed assignTo, status, deadline, compound |
| notifications | 3 | 1 | Removed createdAt (sorting not critical) |
| calendar_events | 1 | 2 | Removed date range indexes (scan with filter) |
| leaveRequests | 3 | 1 | Removed compound index |
| activity_logs | 3 | 1 | Removed standalone timestamp |
| project_details | 2 | 0 | All indexes needed |
| user_account_links | 3 | 0 | All indexes needed |
| signupInvitations | 4 | 0 | All indexes needed |
| role_permissions | 3 | 0 | All indexes needed |
| info-portal | 1 | 2 | Removed nested array indexes |

**Total**: 28 indexes created (down from ~40+ if all were included)

## When to Add More Indexes

Add indexes only if:

1. **Query Performance Issue**: Query is slow (>100ms) and frequently executed
2. **Large Result Sets**: Sorting or filtering >1000 documents
3. **Frequent Combined Queries**: Multiple fields queried together frequently

### How to Identify Missing Indexes

1. **Enable MongoDB Profiler**:
   ```javascript
   db.setProfilingLevel(2, { slowms: 100 })
   ```

2. **Check Slow Queries**:
   ```javascript
   db.system.profile.find().sort({ ts: -1 }).limit(10)
   ```

3. **Use explain()**:
   ```javascript
   db.collection.find({ field: "value" }).explain("executionStats")
   ```

4. **Look for**:
   - `COLLSCAN` (collection scan) - indicates missing index
   - High `executionTimeMillis`
   - High `docsExamined` vs `docsReturned`

### Adding Indexes Safely

1. **Test in Development First**
2. **Monitor Write Performance** - Check if writes slow down
3. **Measure Query Improvement** - Ensure index actually helps
4. **Add During Low Traffic** - Index creation can lock collection

## Best Practices

### ✅ Do

- Index fields used in `QueryByIndex()` operations
- Use compound indexes for queries that filter by multiple fields
- Monitor query performance and add indexes as needed
- Test index impact on write performance

### ❌ Don't

- Create indexes "just in case"
- Index every field
- Create indexes for rarely-used queries
- Create indexes for small collections (<1000 docs)
- Create indexes for fields with low cardinality (few unique values)

## Performance Monitoring

### Key Metrics to Watch

1. **Write Latency**: Should remain <10ms for indexed writes
2. **Query Performance**: Indexed queries should be <50ms
3. **Index Size**: Monitor disk usage
4. **Memory Usage**: Indexes consume RAM

### MongoDB Commands

```javascript
// Check index usage
db.collection.aggregate([{ $indexStats: {} }])

// Check index size
db.collection.stats().indexSizes

// List all indexes
db.collection.getIndexes()
```

## Migration Notes

- Indexes are created automatically when collections are first accessed
- This script creates indexes explicitly for consistency
- Index creation is idempotent (safe to run multiple times)
- Existing indexes are not modified

## Summary

By creating only essential indexes, we:
- ✅ Minimize write overhead (~30-40% reduction)
- ✅ Reduce storage requirements
- ✅ Lower memory usage
- ✅ Maintain query performance for critical operations
- ✅ Allow adding indexes later as needed

**Result**: Faster writes, lower costs, same query performance for essential operations.

---

**Last Updated**: 2025-01-XX

