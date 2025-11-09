# DynamoDB Free Tier Guide

## ✅ Yes, You Can Use Free Tier!

DynamoDB offers a **forever free tier** (not limited to 12 months) that is perfect for 10-15 users.

## Free Tier Benefits

### Forever Free (No Expiration):
- **25 GB of storage** - Free forever
- **25 provisioned Write Capacity Units (WCU)** - Free forever  
- **25 provisioned Read Capacity Units (RCU)** - Free forever
- **2.5 million stream read requests** - Free forever

### First 12 Months (On-Demand):
- **25 GB of storage** - Free
- **200 million read requests** - Free
- **200 million write requests** - Free

## Recommended Setup for 10-15 Users

### Option 1: Provisioned Capacity (Best for Free Tier) ✅

**Settings:**
- Base table: 5 RCU, 5 WCU
- Each GSI: 5 RCU, 5 WCU

**Why This Works:**
- 5 RCU = 5 reads/second = 432,000 reads/day
- 5 WCU = 5 writes/second = 432,000 writes/day
- Well within free tier (using only 20% of free allocation)

**Cost: $0/month** (Forever free)

### Option 2: On-Demand (First 12 Months Free)

**First 12 Months:**
- Free tier covers all usage for 10-15 users
- **Cost: $0/month**

**After 12 Months:**
- Pay per request (~$1-2/month for 10-15 users)
- Still very affordable

## Capacity Planning

### For 10-15 Users:

**Estimated Usage:**
- Reads: ~1-2 per second average = ~86,400-172,800/day
- Writes: ~0.5-1 per second average = ~43,200-86,400/day
- Storage: ~100-500 MB (well under 25 GB free tier)

**With 5 RCU/5 WCU:**
- Capacity: 432,000 reads/day, 432,000 writes/day
- **Headroom: 2-5x your actual usage**
- **Safe for growth up to 50-100 users**

## Table Creation with Free Tier

All tables should use:
```bash
--billing-mode PROVISIONED \
--provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
```

Each GSI should also use:
```
ProvisionedThroughput={ReadCapacityUnits=5,WriteCapacityUnits=5}
```

## Cost Comparison

| Setup | Monthly Cost | Notes |
|-------|--------------|-------|
| **Provisioned (Free Tier)** | **$0** ✅ | Forever free, recommended |
| On-Demand (First 12 months) | $0 | Free tier covers usage |
| On-Demand (After 12 months) | ~$1-2 | Still very affordable |

## When to Upgrade

Upgrade from free tier if:
- You exceed 25 GB storage
- You need more than 25 RCU/WCU
- You have 100+ concurrent users
- You have high traffic spikes

For 10-15 users, **free tier is more than sufficient!**

## Summary

✅ **Yes, use DynamoDB free tier!**
- Use **provisioned capacity** with 5 RCU/5 WCU
- **Cost: $0/month forever**
- Plenty of capacity for 10-15 users
- Room to grow to 50-100 users

See `DYNAMODB_TABLES.md` for table creation commands with free tier settings.

