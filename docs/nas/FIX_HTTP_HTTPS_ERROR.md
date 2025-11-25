# Fix: "HTTP response to HTTPS client" Error

This error occurs when you're trying to connect using HTTPS (`https://`) but MinIO is actually running on HTTP.

## Quick Fix

**Change from HTTPS to HTTP:**

```bash
# Wrong (causes error):
mc alias set myminio https://localhost:9000 minioadmin minioadmin

# Correct:
mc alias set myminio http://localhost:9000 minioadmin minioadmin
```

## Why This Happens

When connecting from **inside the container** (using Container Station Execute or Terminal):

- MinIO listens on HTTP internally (`http://localhost:9000`)
- SSL/HTTPS is typically configured for **external access** only
- Even if SSL certificates are mounted, the internal connection uses HTTP

## When to Use HTTP vs HTTPS

### Use HTTP (`http://`) when:
- ✅ Connecting from **inside the container** (Container Station Execute/Terminal)
- ✅ Using `localhost:9000` or `127.0.0.1:9000`
- ✅ Running commands via Container Station

### Use HTTPS (`https://`) when:
- ✅ Connecting from **outside the container** (your Windows computer)
- ✅ Using your NAS's external IP address
- ✅ SSL certificates are properly configured

## Complete Command (Fixed)

### Using Execute Button:

```bash
wget -q -O /tmp/mc https://dl.min.io/client/mc/release/linux-amd64/mc && chmod +x /tmp/mc && /tmp/mc alias set myminio http://localhost:9000 minioadmin minioadmin && /tmp/mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
```

**Note:** Notice `http://localhost:9000` (not `https://`)

### Using Terminal:

```bash
mc alias set myminio http://localhost:9000 minioadmin minioadmin
mc admin user svcacct add myminio --name ar-13-backend-key --policy readwrite
```

## Related Errors

### "HTTPS response to HTTP client"

**Opposite problem:** MinIO is configured for SSL but you're using HTTP.

**Solution:** Use `https://` instead:
```bash
mc alias set myminio https://localhost:9000 minioadmin minioadmin
```

### "Unable to initialize new alias"

Could be:
1. Wrong protocol (HTTP vs HTTPS) - try the opposite
2. Wrong credentials - check MINIO_ROOT_USER and MINIO_ROOT_PASSWORD
3. MinIO not running - check container status

## Testing the Connection

After setting the alias, test it:

```bash
# List buckets (should work if connection is correct)
mc ls myminio

# If you get errors, try:
# 1. Check if using http:// (not https://) for localhost
# 2. Verify credentials match environment variables
# 3. Check container is running
```

## Summary

**For Container Station (Execute/Terminal):**
- ✅ Use: `http://localhost:9000`
- ❌ Don't use: `https://localhost:9000`

**For Windows Computer:**
- ✅ Use: `https://YOUR-NAS-IP:9000` (if SSL configured)
- ✅ Or: `http://YOUR-NAS-IP:9000` (if SSL not configured)

