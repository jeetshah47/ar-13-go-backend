# Quick Guide: Create Access Keys Using Container Station Terminal

This is the **easiest method** if you have Container Station terminal access.

## Steps

1. **Open Container Station** on your NAS

2. **Find your MinIO container** (named `minio`)

3. **Click on the container** → **Click "Attach Terminal"** tab

4. **If you see a message about enabling interactive/TTY:**
   - The docker-compose file needs to be updated with `stdin_open: true` and `tty: true`
   - Stop the container, update the docker-compose file, and restart
   - Or click the "Attach" button if available

5. **Once terminal opens**, you'll see a command prompt

4. **Install MinIO Client** (one-time setup):
   ```bash
   # Download MinIO client
   wget https://dl.min.io/client/mc/release/linux-amd64/mc
   chmod +x mc
   mv mc /usr/local/bin/mc
   ```
   
   **If `wget` is not available, use `curl`:**
   ```bash
   curl https://dl.min.io/client/mc/release/linux-amd64/mc -o mc
   chmod +x mc
   mv mc /usr/local/bin/mc
   ```

5. **Configure MinIO connection:**
   ```bash
   # Use localhost since we're inside the container
   # IMPORTANT: Use http:// (not https://) for localhost connections
   # Replace 'minioadmin' and 'minioadmin' with your actual credentials
   # from docker-compose.minio.nas.yml (MINIO_ROOT_USER and MINIO_ROOT_PASSWORD)
   mc alias set myminio http://localhost:9000 minioadmin minioadmin
   ```
   
   **Example if your credentials are different:**
   ```bash
   mc alias set myminio http://localhost:9000 your-username your-password
   ```
   
   **Note:** Even if SSL is configured for external access, use `http://` when connecting to `localhost:9000` from inside the container.

6. **Create the access key:**
   ```bash
   mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
   ```

7. **You'll see output like this:**
   ```
   Access Key: ABCDEFGHIJKLMNOP
   Secret Key: abcdefghijklmnopqrstuvwxyz1234567890
   ```
   
   **⚠️ IMPORTANT:** Copy both keys immediately! You won't see the Secret Key again.

8. **Add to your backend `.env` file:**
   ```env
   MINIO_ENDPOINT=your-nas-domain:9000
   MINIO_ACCESS_KEY=ABCDEFGHIJKLMNOP
   MINIO_SECRET_KEY=abcdefghijklmnopqrstuvwxyz1234567890
   MINIO_BUCKET=ar-13-uploads
   MINIO_USE_SSL=true
   MINIO_INSECURE_SSL=false
   ```

## Verify It Works

Test the connection:
```bash
# List buckets
mc ls myminio

# Check if your bucket exists
mc ls myminio/ar-13-uploads
```

## Troubleshooting

### Error: "command not found: mc"
- Make sure you installed it correctly (step 4)
- Try: `which mc` to check if it's in PATH
- If not found, use full path: `/usr/local/bin/mc alias set ...`

### Error: "Unable to parse server response"
- Check your credentials (MINIO_ROOT_USER and MINIO_ROOT_PASSWORD)
- Make sure you're using `http://localhost:9000` (not https) when connecting from inside container

**Error: "HTTP response to HTTPS client"**
- You're using `https://` but MinIO is running on HTTP
- **Solution:** Change to `http://localhost:9000`
- Even if SSL is configured externally, use HTTP for localhost connections

### Error: "Access Denied"
- Make sure you're using root credentials (MINIO_ROOT_USER/MINIO_ROOT_PASSWORD)
- Regular access keys don't have admin privileges

### Error: "dial tcp: i/o timeout" (When connecting from your computer)

If you're trying to connect from your Windows computer (not from Container Station terminal), you need:

1. **Use your NAS's external IP address**, not the container IP (172.29.4.2)
   - Container IP (172.29.4.2) is only accessible from inside the NAS
   - Find your NAS IP: Check your router or NAS settings
   - Common formats: `192.168.1.x` or `192.168.0.x`

2. **Use HTTPS if SSL is enabled:**
   ```bash
   # Correct format (from your computer):
   mc alias set myminio https://YOUR-NAS-IP:9000 minioadmin minioadmin
   
   # Example:
   mc alias set myminio https://192.168.1.100:9000 minioadmin minioadmin
   ```

3. **If using self-signed certificate, add `--insecure` flag:**
   ```bash
   mc alias set myminio https://YOUR-NAS-IP:9000 minioadmin minioadmin --insecure
   ```

4. **Check firewall settings:**
   - Make sure port 9000 is open on your NAS
   - Check if your router/firewall allows connections to port 9000

5. **Verify MinIO is accessible:**
   - Try accessing in browser: `https://YOUR-NAS-IP:9001` (Console)
   - If browser works but mc doesn't, it's likely a certificate issue

**Note:** The easiest way is to use Container Station terminal (as shown in steps above) since it's already inside the NAS network.

## Next Steps

After creating the access key:
1. ✅ Copy Access Key and Secret Key
2. ✅ Add to your `.env` file
3. ✅ Create bucket (if not already created): `mc mb myminio/ar-13-uploads`
4. ✅ Test your backend connection

