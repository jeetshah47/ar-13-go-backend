# Documentation Index

This directory contains all project documentation organized by category.

## 📁 Documentation Structure

### 📡 [API Documentation](./api/)
Complete API reference documentation for all endpoints.

- **[Calendar API](./api/CALENDAR_API.md)** - Calendar event endpoints and Google Calendar integration
- **[Employee API](./api/EMPLOYEE_API.md)** - Employee information, task counts, and statistics
- **[Task Deadline & Progress API](./api/TASK_DEADLINE_PROGRESS_API.md)** - Task deadline and progress tracking endpoints
- **[WebSocket API](./api/WEBSOCKET_API.md)** - Real-time WebSocket communication
- **[WebSocket Client Integration](./api/WEBSOCKET_CLIENT_INTEGRATION.md)** - Client-side WebSocket integration guide

### 🗄️ [Database Documentation](./database/)
DynamoDB setup, configuration, and schema documentation.

- **[DynamoDB Tables Setup](./database/DYNAMODB_TABLES.md)** - Complete table creation guide
- **[Create Missing Tables](./database/CREATE_MISSING_TABLES.md)** - Quick setup for missing tables
- **[DynamoDB Free Tier](./database/DYNAMODB_FREE_TIER.md)** - Free tier information and cost optimization
- **[DynamoDB Field Updates Summary](./database/DYNAMODB_FIELD_UPDATES_SUMMARY.md)** - Summary of all field changes
- **[Model Field Audit](./database/MODEL_FIELD_AUDIT.md)** - Complete audit of all model fields

### 🔄 [Migration Guides](./migration/)
Migration documentation from Firebase to DynamoDB and field updates.

- **[Migration Guide](./migration/MIGRATION_GUIDE.md)** - Complete migration guide from Firebase to DynamoDB
- **[Migration from Backup](./migration/MIGRATION_FROM_BACKUP.md)** - How to migrate data from backup files
- **[Task Field Migration](./migration/TASK_FIELD_MIGRATION.md)** - Task field migration (duration → deadline)
- **[Firebase Removal](./migration/FIREBASE_REMOVAL_COMPLETE.md)** - Firebase removal completion status
- **[Firebase Removal Summary](./migration/FIREBASE_REMOVAL_SUMMARY.md)** - Summary of Firebase removal
- **[Migration Complete](./migration/MIGRATION_COMPLETE.md)** - Migration completion status
- **[Migration Status](./migration/MIGRATION_STATUS.md)** - Current migration status
- **[Migration Summary](./migration/MIGRATION_SUMMARY.md)** - Migration summary
- **[Repository Migration Status](./migration/REPOSITORY_MIGRATION_STATUS.md)** - Repository migration tracking
- **[Remove Firebase Script](./migration/REMOVE_FIREBASE_SCRIPT.md)** - Script documentation for Firebase removal

### ⚙️ [Setup & Configuration](./setup/)
Getting started guides and setup instructions.

- **[Quick Start Guide](./setup/QUICK_START.md)** - Get started quickly with the project
- **[Dependencies](./setup/DEPENDENCIES.md)** - Project dependencies and requirements
- **[Testing Guide](./setup/TESTING_GUIDE.md)** - How to test the application

### 💻 [Development Documentation](./development/)
Development progress, changes, and next steps.

- **[Route Changes](./development/ROUTE_CHANGES.md)** - API route changes and updates
- **[Progress Summary](./development/PROGRESS_SUMMARY.md)** - Development progress tracking
- **[Next Steps Completed](./development/NEXT_STEPS_COMPLETED.md)** - Completed development tasks

## 🚀 Quick Navigation

### For New Developers
1. Start with [Quick Start Guide](./setup/QUICK_START.md)
2. Review [Dependencies](./setup/DEPENDENCIES.md)
3. Check [DynamoDB Tables Setup](./database/DYNAMODB_TABLES.md)
4. Explore [API Documentation](./api/)

### For API Integration
1. Review [API Documentation](./api/) for endpoint details
2. Check [Route Changes](./development/ROUTE_CHANGES.md) for latest updates
3. See [WebSocket Client Integration](./api/WEBSOCKET_CLIENT_INTEGRATION.md) for real-time features

### For Database Setup
1. Read [DynamoDB Tables Setup](./database/DYNAMODB_TABLES.md)
2. Check [Create Missing Tables](./database/CREATE_MISSING_TABLES.md) if needed
3. Review [DynamoDB Free Tier](./database/DYNAMODB_FREE_TIER.md) for cost optimization

### For Migration Tasks
1. Start with [Migration Guide](./migration/MIGRATION_GUIDE.md)
2. Check [Task Field Migration](./migration/TASK_FIELD_MIGRATION.md) for field updates
3. Review [Migration Status](./migration/MIGRATION_STATUS.md) for current state

## 📝 Documentation Standards

- All API documentation follows OpenAPI-style format
- Migration guides include step-by-step instructions
- Database documentation includes schema and setup commands
- Code examples are provided where applicable

## 🔍 Finding Documentation

- **API Endpoints**: Check `api/` folder
- **Database Setup**: Check `database/` folder
- **Migration Help**: Check `migration/` folder
- **Getting Started**: Check `setup/` folder
- **Development Info**: Check `development/` folder

## 📚 Additional Resources

- Main project README: [../README.md](../README.md)
- Source code: [../internal/](../internal/)
- Scripts: [../scripts/](../scripts/)

---

**Last Updated**: 2025-01-XX  
**Documentation Version**: 1.0

