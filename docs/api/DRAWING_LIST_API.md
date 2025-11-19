# Drawing List Master Data API Documentation

Complete documentation for the Drawing List Master Data module - Architecture drawing categories and types.

## Overview

The Drawing List module provides CRUD APIs for managing architectural drawing categories and their associated drawing types. This is a master data module that stores default task list items for architecture drawings.

## Database Collections

- **drawing_categories**: Stores drawing categories (e.g., "Architectural Drawings", "Structural Drawings")
- **drawing_types**: Stores specific drawing types under each category (e.g., "Floor Plans", "Elevations")

## Data Model

### DrawingCategory
```json
{
  "id": "string",
  "name": "string",
  "description": "string (optional)",
  "order": "number",
  "isActive": "boolean",
  "created": "ISO 8601 datetime",
  "updated": "ISO 8601 datetime (optional)"
}
```

### DrawingType
```json
{
  "id": "string",
  "categoryId": "string",
  "name": "string",
  "description": "string (optional)",
  "order": "number",
  "isActive": "boolean",
  "created": "ISO 8601 datetime",
  "updated": "ISO 8601 datetime (optional)"
}
```

## API Endpoints

All endpoints require authentication and appropriate permissions.

### Base Path
```
/api/drawing-list
```

### Category Endpoints

#### Get All Categories
```
GET /api/drawing-list/categories
```

**Response:**
```json
{
  "categories": [
    {
      "id": "uuid",
      "name": "Architectural Drawings",
      "description": "Technical drawings of buildings and structures",
      "order": 1,
      "isActive": true,
      "created": "2025-01-20T10:00:00Z",
      "updated": null
    }
  ],
  "total": 1
}
```

**Permissions Required:** `drawingList:read`

#### Get Category by ID
```
GET /api/drawing-list/categories/:id
```

**Response:**
```json
{
  "category": {
    "id": "uuid",
    "name": "Architectural Drawings",
    "description": "Technical drawings of buildings and structures",
    "order": 1,
    "isActive": true,
    "created": "2025-01-20T10:00:00Z",
    "updated": null
  }
}
```

**Permissions Required:** `drawingList:read`

#### Create Category
```
POST /api/drawing-list/categories
```

**Request Body:**
```json
{
  "name": "Architectural Drawings",
  "description": "Technical drawings of buildings and structures",
  "order": 1,
  "isActive": true
}
```

**Response:**
```json
{
  "message": "Category created successfully",
  "category": {
    "id": "uuid",
    "name": "Architectural Drawings",
    "description": "Technical drawings of buildings and structures",
    "order": 1,
    "isActive": true,
    "created": "2025-01-20T10:00:00Z",
    "updated": null
  }
}
```

**Permissions Required:** `drawingList:write`

#### Update Category
```
PUT /api/drawing-list/categories/:id
```

**Request Body:**
```json
{
  "name": "Architectural Drawings",
  "description": "Updated description",
  "order": 1,
  "isActive": true
}
```

**Response:**
```json
{
  "message": "Category updated successfully",
  "category": {
    "id": "uuid",
    "name": "Architectural Drawings",
    "description": "Updated description",
    "order": 1,
    "isActive": true,
    "created": "2025-01-20T10:00:00Z",
    "updated": "2025-01-20T11:00:00Z"
  }
}
```

**Permissions Required:** `drawingList:write`

#### Delete Category
```
DELETE /api/drawing-list/categories/:id
```

**Response:**
```json
{
  "message": "Category deleted successfully"
}
```

**Note:** Cannot delete a category if it has associated drawing types.

**Permissions Required:** `drawingList:delete`

### Type Endpoints

#### Get All Types
```
GET /api/drawing-list/types
```

**Response:**
```json
{
  "types": [
    {
      "id": "uuid",
      "categoryId": "uuid",
      "name": "Floor Plans",
      "description": "Horizontal cross-section views of buildings",
      "order": 1,
      "isActive": true,
      "created": "2025-01-20T10:00:00Z",
      "updated": null
    }
  ],
  "total": 1
}
```

**Permissions Required:** `drawingList:read`

#### Get Type by ID
```
GET /api/drawing-list/types/:id
```

**Response:**
```json
{
  "type": {
    "id": "uuid",
    "categoryId": "uuid",
    "name": "Floor Plans",
    "description": "Horizontal cross-section views of buildings",
    "order": 1,
    "isActive": true,
    "created": "2025-01-20T10:00:00Z",
    "updated": null
  }
}
```

**Permissions Required:** `drawingList:read`

#### Get Types by Category
```
GET /api/drawing-list/types/category/:categoryId
```

**Response:**
```json
{
  "types": [
    {
      "id": "uuid",
      "categoryId": "uuid",
      "name": "Floor Plans",
      "description": "Horizontal cross-section views of buildings",
      "order": 1,
      "isActive": true,
      "created": "2025-01-20T10:00:00Z",
      "updated": null
    }
  ],
  "total": 1
}
```

**Permissions Required:** `drawingList:read`

#### Create Type
```
POST /api/drawing-list/types
```

**Request Body:**
```json
{
  "categoryId": "uuid",
  "name": "Floor Plans",
  "description": "Horizontal cross-section views of buildings",
  "order": 1,
  "isActive": true
}
```

**Response:**
```json
{
  "message": "Drawing type created successfully",
  "type": {
    "id": "uuid",
    "categoryId": "uuid",
    "name": "Floor Plans",
    "description": "Horizontal cross-section views of buildings",
    "order": 1,
    "isActive": true,
    "created": "2025-01-20T10:00:00Z",
    "updated": null
  }
}
```

**Permissions Required:** `drawingList:write`

#### Update Type
```
PUT /api/drawing-list/types/:id
```

**Request Body:**
```json
{
  "categoryId": "uuid",
  "name": "Floor Plans",
  "description": "Updated description",
  "order": 1,
  "isActive": true
}
```

**Response:**
```json
{
  "message": "Drawing type updated successfully",
  "type": {
    "id": "uuid",
    "categoryId": "uuid",
    "name": "Floor Plans",
    "description": "Updated description",
    "order": 1,
    "isActive": true,
    "created": "2025-01-20T10:00:00Z",
    "updated": "2025-01-20T11:00:00Z"
  }
}
```

**Permissions Required:** `drawingList:write`

#### Delete Type
```
DELETE /api/drawing-list/types/:id
```

**Response:**
```json
{
  "message": "Drawing type deleted successfully"
}
```

**Permissions Required:** `drawingList:delete`

### Combined Endpoints

#### Get Categories with Types
```
GET /api/drawing-list/categories-with-types
```

**Response:**
```json
{
  "categories": [
    {
      "id": "uuid",
      "name": "Architectural Drawings",
      "description": "Technical drawings of buildings and structures",
      "order": 1,
      "isActive": true,
      "created": "2025-01-20T10:00:00Z",
      "updated": null,
      "types": [
        {
          "id": "uuid",
          "categoryId": "uuid",
          "name": "Floor Plans",
          "description": "Horizontal cross-section views of buildings",
          "order": 1,
          "isActive": true,
          "created": "2025-01-20T10:00:00Z",
          "updated": null
        }
      ]
    }
  ],
  "total": 1
}
```

**Permissions Required:** `drawingList:read`

## Seeding Initial Data

To seed the database with initial drawing categories and types:

```bash
go run scripts/seed_drawing_list.go
```

To force update existing data:

```bash
go run scripts/seed_drawing_list.go -force
```

The seeding script will create:
- 4 default categories (Architectural Drawings, Structural Drawings, MEP Drawings, Site Plans)
- 12 default types distributed across categories

## Error Responses

### 400 Bad Request
```json
{
  "error": "category name is required"
}
```

### 404 Not Found
```json
{
  "error": "Category not found"
}
```

### 403 Forbidden
```json
{
  "error": "insufficient permissions"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error message"
}
```

## Validation Rules

### Category
- `name` is required
- `order` should be a positive integer
- `isActive` defaults to `true`

### Type
- `name` is required
- `categoryId` is required and must reference an existing category
- `order` should be a positive integer
- `isActive` defaults to `true`

## Business Rules

1. **Category Deletion**: A category cannot be deleted if it has associated drawing types
2. **Type Category Reference**: When creating or updating a type, the `categoryId` must reference an existing category
3. **Ordering**: Categories and types are ordered by their `order` field (ascending)

## Permissions

The following permissions are required:
- `drawingList:read` - Read access to categories and types
- `drawingList:write` - Create and update access
- `drawingList:delete` - Delete access

Make sure these permissions are configured in your role-based access control system.

