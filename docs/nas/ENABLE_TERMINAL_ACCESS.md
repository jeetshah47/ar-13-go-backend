# Enable Terminal Access in Container Station

If you see the message: **"You must enable the interactive (-i) and TTY (-t) processes to attach the terminal to a container"**, follow these steps:

## Solution: Update docker-compose File

The docker-compose file needs to include `stdin_open: true` and `tty: true` to enable terminal access.

### Step 1: Stop the Container

1. In Container Station, find your MinIO container
2. Click **Stop** button
3. Wait for it to stop completely

### Step 2: Update docker-compose File

Edit `docker-compose.minio.nas.yml` and add these lines under the `minio` service:

```yaml
services:
  minio:
    # ... existing configuration ...
    stdin_open: true
    tty: true
```

**Full example:**
```yaml
version: '3.8'

services:
  minio:
    image: minio/minio:latest
    container_name: minio
    restart: unless-stopped
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
      TZ: Asia/Kolkata
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - /share/Container/minio/data:/data
      - /share/Container/minio/config:/root/.minio
      - /share/Container/minio/certs:/root/.minio/certs
    networks:
      - minio-network
    stdin_open: true    # ← Add this
    tty: true           # ← Add this
    healthcheck:
      test: ["CMD", "curl", "-f", "https://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
      start_period: 40s

networks:
  minio-network:
    driver: bridge
```

### Step 3: Restart the Container

**Option A: Using Container Station UI**
1. Click **Edit** on the container
2. Update the YAML configuration
3. Click **Apply** or **Update**
4. Start the container

**Option B: Using SSH**
1. SSH into your NAS
2. Navigate to where your docker-compose file is:
   ```bash
   cd /share/Container/minio
   ```
3. Restart with updated configuration:
   ```bash
   docker-compose -f docker/docker-compose.minio.nas.yml down
   docker-compose -f docker/docker-compose.minio.nas.yml up -d
   ```

### Step 4: Verify Terminal Access

1. Go back to Container Station
2. Click on your MinIO container
3. Click **Attach Terminal** tab
4. You should now see a terminal prompt instead of the error message

## Alternative: Use "Execute" Command

If you can't update the docker-compose file right now, you can use the **Execute** button in Container Station:

1. Click on your MinIO container
2. Click **Execute** button
3. Enter a command like:
   ```bash
   /bin/sh
   ```
4. This will open a shell session

However, this is temporary. For permanent terminal access, update the docker-compose file as shown above.

## What These Options Do

- **`stdin_open: true`** - Enables interactive mode (`-i` flag)
- **`tty: true`** - Allocates a pseudo-TTY (`-t` flag)

These are required for Container Station's terminal attachment feature to work.

