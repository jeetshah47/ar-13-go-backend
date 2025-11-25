# SSL Certificate Setup for MinIO on NAS

This guide explains how to set up SSL certificates for MinIO on your NAS.

## Certificate Requirements

MinIO requires SSL certificates in a specific format:
- **Public Certificate:** `public.crt` (or `CERTIFICATE.crt`)
- **Private Key:** `private.key` (or `PRIVATE.key`)

The certificates must be placed in: `/share/Container/minio/certs/`

## Option 1: Using Let's Encrypt (Recommended for Production)

If you have a domain name pointing to your NAS:

1. **Install Certbot on your NAS** (if available) or use a reverse proxy like Nginx with Let's Encrypt

2. **Obtain certificates:**
   ```bash
   certbot certonly --standalone -d your-nas-domain.com
   ```

3. **Copy certificates to MinIO directory:**
   ```bash
   cp /etc/letsencrypt/live/your-nas-domain.com/fullchain.pem /share/Container/minio/certs/public.crt
   cp /etc/letsencrypt/live/your-nas-domain.com/privkey.pem /share/Container/minio/certs/private.key
   ```

4. **Set permissions:**
   ```bash
   chmod 600 /share/Container/minio/certs/private.key
   chmod 644 /share/Container/minio/certs/public.crt
   ```

5. **Set up auto-renewal** (if using certbot):
   ```bash
   # Add to crontab to renew and copy certificates
   0 0 * * * certbot renew && cp /etc/letsencrypt/live/your-nas-domain.com/fullchain.pem /share/Container/minio/certs/public.crt && cp /etc/letsencrypt/live/your-nas-domain.com/privkey.pem /share/Container/minio/certs/private.key && docker restart minio
   ```

## Option 2: Using Existing Certificates

If you already have SSL certificates:

1. **Copy your certificates:**
   ```bash
   cp /path/to/your/certificate.crt /share/Container/minio/certs/public.crt
   cp /path/to/your/private.key /share/Container/minio/certs/private.key
   ```

2. **Set permissions:**
   ```bash
   chmod 600 /share/Container/minio/certs/private.key
   chmod 644 /share/Container/minio/certs/public.crt
   ```

3. **Verify certificate format:**
   ```bash
   openssl x509 -in /share/Container/minio/certs/public.crt -text -noout
   ```

## Option 3: Generate Self-Signed Certificates (Development/Testing)

For testing or internal use only:

1. **SSH into your NAS** or use a machine with OpenSSL

2. **Create certificates directory:**
   ```bash
   mkdir -p /share/Container/minio/certs
   cd /share/Container/minio/certs
   ```

3. **Generate private key:**
   ```bash
   openssl genrsa -out private.key 2048
   ```

4. **Generate certificate signing request:**
   ```bash
   openssl req -new -key private.key -out cert.csr \
     -subj "/C=US/ST=State/L=City/O=Organization/CN=your-nas-domain.com"
   ```
   
   Replace the values:
   - `C`: Country code (e.g., US, IN)
   - `ST`: State/Province
   - `L`: City/Locality
   - `O`: Organization name
   - `CN`: Common Name (your NAS domain or IP)

5. **Generate self-signed certificate:**
   ```bash
   openssl x509 -req -days 365 -in cert.csr -signkey private.key -out public.crt
   ```

6. **Set permissions:**
   ```bash
   chmod 600 private.key
   chmod 644 public.crt
   ```

7. **Clean up:**
   ```bash
   rm cert.csr
   ```

8. **Verify certificate:**
   ```bash
   openssl x509 -in public.crt -text -noout
   ```

## Certificate for Multiple Domains (SAN Certificate)

If you need to access MinIO via multiple domains or IPs:

1. **Create a certificate configuration file** `cert.conf`:
   ```ini
   [req]
   default_bits = 2048
   distinguished_name = req_distinguished_name
   req_extensions = v3_req
   prompt = no

   [req_distinguished_name]
   C = US
   ST = State
   L = City
   O = Organization
   CN = your-nas-domain.com

   [v3_req]
   keyUsage = keyEncipherment, dataEncipherment
   extendedKeyUsage = serverAuth
   subjectAltName = @alt_names

   [alt_names]
   DNS.1 = your-nas-domain.com
   DNS.2 = *.your-nas-domain.com
   IP.1 = 192.168.1.100
   IP.2 = 10.0.0.50
   ```

2. **Generate certificate:**
   ```bash
   openssl req -new -x509 -key private.key -out public.crt -days 365 -config cert.conf -extensions v3_req
   ```

## Verify SSL Setup

After setting up certificates and starting MinIO:

1. **Check if MinIO is using SSL:**
   ```bash
   docker logs minio | grep -i ssl
   ```

2. **Test HTTPS connection:**
   ```bash
   curl -k https://your-nas-ip:9000/minio/health/live
   ```

3. **Access Console via HTTPS:**
   - Open: `https://your-nas-ip:9001`
   - If using self-signed certificate, accept the security warning

## Troubleshooting

### Certificate Not Found Error

**Error:** `Unable to load TLS certificate`

**Solution:**
- Ensure certificates are in `/share/Container/minio/certs/`
- Check file names: `public.crt` and `private.key`
- Verify file permissions

### Permission Denied

**Error:** `Permission denied` when accessing private key

**Solution:**
```bash
chmod 600 /share/Container/minio/certs/private.key
```

### Certificate Format Error

**Error:** `Invalid certificate format`

**Solution:**
- Ensure certificate is in PEM format (not DER)
- Check certificate content starts with `-----BEGIN CERTIFICATE-----`
- Verify certificate is not corrupted

### Self-Signed Certificate Warnings

**Warning:** Browser shows "Not Secure" or certificate warning

**Solution:**
- This is normal for self-signed certificates
- For production, use Let's Encrypt or a trusted CA
- In your backend, set `MINIO_INSECURE_SSL=true` if you want to skip verification

## Certificate Renewal

For Let's Encrypt certificates (valid for 90 days):

1. **Set up auto-renewal** (see Option 1)
2. **Manual renewal:**
   ```bash
   certbot renew
   cp /etc/letsencrypt/live/your-nas-domain.com/fullchain.pem /share/Container/minio/certs/public.crt
   cp /etc/letsencrypt/live/your-nas-domain.com/privkey.pem /share/Container/minio/certs/private.key
   docker restart minio
   ```

## Security Best Practices

1. **Protect Private Key:**
   - Never share or commit private key to version control
   - Use `chmod 600` for private key
   - Backup securely

2. **Use Strong Certificates:**
   - Minimum 2048-bit RSA key
   - Use SHA-256 or better
   - Keep certificates up to date

3. **Regular Renewal:**
   - Set up automatic renewal for Let's Encrypt
   - Monitor certificate expiration dates

4. **Certificate Validation:**
   - In production, always use certificates from trusted CAs
   - Avoid self-signed certificates for production

