# Environment Variables Setup for NAS Deployment

This guide explains how to set up the `.env` file for the AR-13 backend on your NAS server.

## Location

Create a `.env` file in the `docker/` directory on your NAS server:
```
/path/to/ar-13-go-backend/docker/.env
```

## Required Environment Variables

### Server Configuration
```bash
PORT=3000
NODE_ENV=production
```

### MongoDB (REQUIRED)
```bash
MONGODB_URI=mongodb://username:password@host:27017/database?authSource=admin
# Or for MongoDB Atlas:
# MONGODB_URI=mongodb+srv://username:password@cluster.mongodb.net/database
MONGODB_DATABASE=ar13_backend
```

### JWT (REQUIRED)
```bash
# Generate a strong secret: openssl rand -base64 32
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRATION_HOURS=24
REFRESH_EXPIRATION_DAYS=30
```

### Redis
```bash
REDIS_ADDR=redis:6379
REDIS_PASSWORD=your-redis-password-here
REDIS_DB=0
REDIS_PORT=6379
```

### Email (Optional but recommended)
```bash
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USER=your-email@gmail.com
EMAIL_PASSWORD=your-app-specific-password
EMAIL_FROM=your-email@gmail.com
EMAIL_FROM_NAME=AR-13 System
```

### Frontend
```bash
FRONTEND_URL=http://your-nas-ip:3000
```

### Google OAuth (Optional)
```bash
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
```

### MinIO/NAS Storage (Optional)
```bash
MINIO_ENDPOINT=minio.example.com
MINIO_ACCESS_KEY=your-minio-access-key
MINIO_SECRET_KEY=your-minio-secret-key
MINIO_BUCKET=ar-13-uploads
MINIO_USE_SSL=false
MINIO_INSECURE_SSL=false
```

### FileBrowser (Optional)
```bash
FILEBROWSER_ENABLED=false
FILEBROWSER_URL=http://localhost:8080
FILEBROWSER_TOKEN=
FILEBROWSER_SERVICE_URL=http://localhost:8082
```

### QNAP NAS Integration (Optional)
```bash
QNAP_NAS_IP=192.168.1.100
QNAP_API_PORT=8080
QNAP_SERVICE_USER=your-qnap-username
QNAP_SERVICE_PASSWORD=your-qnap-password
QNAP_SHARE_NAME=studio-work
QNAP_SESSION_TIMEOUT=30
```

## Setup Instructions

### Option 1: SSH into NAS and create the file

1. SSH into your NAS server
2. Navigate to the project directory:
   ```bash
   cd /path/to/ar-13-go-backend/docker
   ```
3. Create the `.env` file:
   ```bash
   nano .env
   # or
   vi .env
   ```
4. Copy and paste the environment variables above, filling in your actual values
5. Save and exit

### Option 2: Use NAS File Manager

1. Open your NAS web interface
2. Navigate to File Station or File Manager
3. Go to the `docker/` directory in your project
4. Create a new file named `.env`
5. Edit it and add all the environment variables

### Option 3: Copy from a template

1. On your NAS server, navigate to the docker directory:
   ```bash
   cd /path/to/ar-13-go-backend/docker
   ```
2. Create the `.env` file from a template:
   ```bash
   cat > .env << 'EOF'
   PORT=3000
   NODE_ENV=production
   MONGODB_URI=your-mongodb-uri
   MONGODB_DATABASE=ar13_backend
   JWT_SECRET=your-jwt-secret
   REDIS_ADDR=redis:6379
   REDIS_PASSWORD=your-redis-password
   # ... add other variables
   EOF
   ```
3. Edit the file to fill in your actual values:
   ```bash
   nano .env
   ```

## Security Notes

- **NEVER commit the `.env` file to git** - it contains sensitive information
- The `.env` file should already be in `.gitignore`
- Use strong, unique passwords and secrets
- For JWT_SECRET, generate a secure random string:
  ```bash
  openssl rand -base64 32
  ```

## Verification

After creating the `.env` file, verify it's in the correct location:
```bash
cd /path/to/ar-13-go-backend/docker
ls -la .env
```

The file should exist and be readable. Then test the deployment:
```bash
docker compose -f docker-compose.nas.yml config
```

This will show you the resolved configuration and help identify any issues.

