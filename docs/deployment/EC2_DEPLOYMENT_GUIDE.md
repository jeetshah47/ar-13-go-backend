# EC2 Deployment Guide - Step by Step

## Prerequisites

- AWS Account
- Domain name (optional, but recommended)
- SSH key pair
- Your Go backend code ready

---

## Step 1: Launch EC2 Instance

### 1.1 Go to EC2 Console
1. Log into AWS Console
2. Navigate to **EC2** → **Instances**
3. Click **Launch Instance**

### 1.2 Configure Instance

**Name:** `ar-13-backend` (or your preferred name)

**AMI (Amazon Machine Image):**
- Select **Ubuntu Server 22.04 LTS** (free tier eligible)
- Architecture: `64-bit (x86)`

**Instance Type:**
- Select **t2.micro** (free tier eligible)
- 1 vCPU, 1 GB RAM

**Key Pair:**
- Create new key pair or select existing
- **Important:** Download the `.pem` file and keep it safe!
- Name: `ar-13-backend-key`

**Network Settings:**
- **VPC:** Default VPC
- **Subnet:** Any availability zone
- **Auto-assign Public IP:** Enable
- **Security Group:** Create new security group
  - Name: `ar-13-backend-sg`
  - Description: `Security group for AR-13 backend`
  - **Inbound Rules:**
    - SSH (22) - My IP (or 0.0.0.0/0 for testing, restrict later)
    - HTTP (80) - 0.0.0.0/0
    - HTTPS (443) - 0.0.0.0/0
    - Custom TCP (3000) - 0.0.0.0/0 (for direct access, optional)

**Storage:**
- **Volume Type:** gp3
- **Size:** 20 GB (free tier includes 30 GB)
- **Delete on Termination:** Uncheck if you want to keep data

**Advanced Details (Optional):**
- Add user data script (see below)

### 1.3 Launch Instance
- Click **Launch Instance**
- Wait for instance to be in "Running" state
- Note the **Public IPv4 address** (e.g., `54.123.45.67`)

---

## Step 2: Connect to EC2 Instance

### 2.1 Set Permissions for Key File
```bash
chmod 400 ar-13-backend-key.pem
```

### 2.2 SSH into Instance
```bash
ssh -i ar-13-backend-key.pem ubuntu@YOUR_EC2_PUBLIC_IP
```

Replace `YOUR_EC2_PUBLIC_IP` with your actual IP address.

---

## Step 3: Initial Server Setup

### 3.1 Update System
```bash
sudo apt update && sudo apt upgrade -y
```

### 3.2 Install Required Software
```bash
# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version  # Verify installation

# Install Nginx
sudo apt install -y nginx

# Install Certbot for SSL
sudo apt install -y certbot python3-certbot-nginx

# Install Git
sudo apt install -y git

# Install Redis (optional, for caching)
sudo apt install -y redis-server
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

### 3.3 Install AWS CLI (for S3 file uploads)
```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
sudo apt install -y unzip
unzip awscliv2.zip
sudo ./aws/install
aws --version
```

---

## Step 4: Deploy Your Application

### 4.1 Clone Your Repository
```bash
cd ~
git clone https://github.com/your-username/ar-13-go-backend.git
cd ar-13-go-backend
```

**OR** if using SSH:
```bash
git clone git@github.com:your-username/ar-13-go-backend.git
```

### 4.2 Build the Application
```bash
cd ~/ar-13-go-backend
go mod download
go build -o server cmd/server/main.go
```

### 4.3 Create Environment File
```bash
nano ~/ar-13-go-backend/.env
```

Add your environment variables:
```env
PORT=3000
NODE_ENV=production

# AWS DynamoDB
AWS_REGION=us-east-1

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production-min-32-chars
JWT_EXPIRATION_HOURS=24
REFRESH_EXPIRATION_DAYS=30

# Email Configuration
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USER=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM=noreply@yourdomain.com
EMAIL_FROM_NAME=AR-13

# Frontend URL
FRONTEND_URL=https://your-frontend-domain.com

# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

**Important Notes:**
- Use strong `JWT_SECRET` (at least 32 characters)
- For Gmail, use App Password (not regular password)
- Update `FRONTEND_URL` with your actual frontend domain
- AWS credentials will be handled via IAM role (see Step 5)

### 4.4 Set Up AWS Credentials

**Option A: IAM Role (Recommended for EC2)**
1. Go to EC2 → Instances → Select your instance
2. Actions → Security → Modify IAM role
3. Create new IAM role with DynamoDB and S3 permissions
4. Attach role to instance

**Option B: Environment Variables (Less Secure)**
```bash
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
```

---

## Step 5: Create Systemd Service

### 5.1 Create Service File
```bash
sudo nano /etc/systemd/system/ar13-backend.service
```

### 5.2 Add Service Configuration
```ini
[Unit]
Description=AR-13 Go Backend Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/ar-13-go-backend
EnvironmentFile=/home/ubuntu/ar-13-go-backend/.env
ExecStart=/home/ubuntu/ar-13-go-backend/server
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### 5.3 Enable and Start Service
```bash
sudo systemctl daemon-reload
sudo systemctl enable ar13-backend
sudo systemctl start ar13-backend
sudo systemctl status ar13-backend
```

### 5.4 Check Logs
```bash
# View logs
sudo journalctl -u ar13-backend -f

# View last 100 lines
sudo journalctl -u ar13-backend -n 100
```

---

## Step 6: Configure Nginx Reverse Proxy

### ⚠️ Important: Elastic IP and Nginx

**No changes needed!** Elastic IP is completely transparent to Nginx. Your Nginx configuration stays exactly the same whether you use Elastic IP or regular EC2 IP. Nginx listens on ports 80/443 and proxies to `localhost:3000` - it doesn't care about the external IP address.

---

### 6.1 Create Nginx Configuration
```bash
sudo nano /etc/nginx/sites-available/ar13-backend
```

### 6.2 Add Configuration
```nginx
server {
    listen 80;
    server_name your-domain.com www.your-domain.com;

    # Increase timeouts for WebSocket
    proxy_connect_timeout 7d;
    proxy_send_timeout 7d;
    proxy_read_timeout 7d;

    # WebSocket endpoint
    location /ws {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API endpoints
    location /api {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Static file uploads
    location /uploads {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://localhost:3000;
        access_log off;
    }
}
```

**Replace `your-domain.com` with your actual domain.**

### 6.3 Enable Site
```bash
sudo ln -s /etc/nginx/sites-available/ar13-backend /etc/nginx/sites-enabled/
sudo nginx -t  # Test configuration
sudo systemctl restart nginx
```

---

## Step 7: Set Up SSL with Let's Encrypt

### 7.1 Get SSL Certificate
```bash
sudo certbot --nginx -d your-domain.com -d www.your-domain.com
```

Follow the prompts:
- Enter your email
- Agree to terms
- Choose whether to redirect HTTP to HTTPS (recommended: Yes)

### 7.2 Auto-Renewal (Already Set Up)
Certbot automatically sets up renewal. Test it:
```bash
sudo certbot renew --dry-run
```

---

## Step 8: Configure Domain DNS

### ⚠️ Important: AWS Route 53 Free Tier

**AWS Route 53 does NOT have a free tier for domain services:**
- ❌ Domain registration: ~$12-15/year per domain
- ❌ Hosted zones: ~$0.50/month per hosted zone
- ❌ DNS queries: ~$0.40 per million queries

**Solution:** Use your domain registrar's DNS management (FREE) + AWS Elastic IP (FREE)

---

### Free Domain Options

**Are there free domain services?** Yes, but with limitations:

#### Option 1: Free Subdomains (For Testing/Development)
- **Freenom** (`.tk`, `.ml`, `.ga`, `.cf`, `.gq`) - ⚠️ **Not recommended** (reliability issues, often blocked)
- **GoogieHost** - Free subdomains like `.cu.ma`, `.thats.im` (with hosting)
- **NoIP** - Free dynamic DNS subdomains (requires renewal every 30 days)

**Limitations:**
- ❌ Not professional for production
- ❌ May have reliability issues
- ❌ Some services require periodic renewal
- ❌ Limited DNS control

#### Option 2: First Year Free (With Hosting Plans)
- **Squarespace** - Free `.com` domain for first year with annual plan
- **Wix** - Free domain for first year with Premium plan
- **DreamHost** - Free domain credit with annual hosting plans

**Limitations:**
- ❌ Requires purchasing hosting (not free overall)
- ❌ Standard renewal rates after first year (~$12-15/year)
- ❌ May have restrictions on DNS management

#### Option 3: Low-Cost Domains (Recommended)
**Best value for production:**
- **Namecheap** - ~$8-12/year for `.com` domains (often with free privacy)
- **Google Domains** - ~$12/year for `.com` (now part of Squarespace)
- **Cloudflare** - At-cost pricing (~$8-10/year) + free DNS
- **Porkbun** - ~$8-10/year for `.com` domains

**Recommendation for Production:**
- ✅ **Use a paid domain** (~$8-12/year) - Professional, reliable, full control
- ✅ **Use Cloudflare** for DNS (FREE) - Fast, reliable, includes DDoS protection
- ✅ **Total cost: ~$8-12/year** (less than $1/month)

**For Development/Testing:**
- Use EC2 Public DNS temporarily
- Or use a free subdomain (with limitations)

#### Option 4: Free DNS with Any Domain (Recommended Setup)

**Best approach:** Buy a domain anywhere, use Cloudflare for FREE DNS:

1. **Buy domain** from Namecheap, GoDaddy, or any registrar (~$8-12/year)
2. **Transfer DNS to Cloudflare** (FREE):
   - Sign up at [cloudflare.com](https://cloudflare.com) (free account)
   - Add your domain to Cloudflare
   - Update nameservers at your registrar
   - **Benefits:**
     - ✅ FREE DNS management
     - ✅ FREE DDoS protection
     - ✅ FREE SSL (if using Cloudflare proxy)
     - ✅ Fast global CDN
     - ✅ Full DNS control

**Total Cost:** ~$8-12/year for domain + $0 for DNS = **~$8-12/year total**

---

### 8.1 Understanding EC2 Public DNS vs Elastic IP

**What is EC2 Public DNS?**
AWS automatically assigns a Public DNS name to your EC2 instance, like:
```
ec2-98-86-113-245.compute-1.amazonaws.com
```

**How it works:**
- ✅ Automatically created by AWS (no setup needed)
- ✅ Resolves to your instance's public IP address
- ✅ Can be used to access your API: `http://ec2-98-86-113-245.compute-1.amazonaws.com`
- ❌ **Changes when instance stops/starts** (unless using Elastic IP)
- ❌ Not user-friendly for production
- ❌ Can't get SSL certificate for AWS domain name

**With Elastic IP:**
- ✅ Public DNS becomes **static** (doesn't change)
- ✅ Still not recommended for production (use your own domain)

**Recommendation:** Use your own domain name (e.g., `api.yourdomain.com`) instead of the AWS Public DNS.

---

### 8.2 Allocate Elastic IP (Recommended - FREE)

**Why Elastic IP?**
- ✅ **FREE** - One Elastic IP per account is free
- ✅ Static IP that doesn't change when instance stops/restarts
- ✅ Makes Public DNS static too
- ✅ Better for production use
- ✅ **Fully supports HTTPS** - Works with SSL certificates (Let's Encrypt)

**Steps:**
1. Go to **EC2 Console** → **Elastic IPs** → **Allocate Elastic IP address**
2. Click **Allocate** (keep default settings)
3. Select the Elastic IP → **Actions** → **Associate Elastic IP address**
4. Select your EC2 instance
5. Click **Associate**

**Note:** Elastic IP is free as long as it's associated with a running instance. If you stop the instance, you may incur charges.

**HTTPS Support:** ✅ Elastic IP fully supports HTTPS. SSL certificates (like Let's Encrypt) are domain-based, not IP-based, so HTTPS works exactly the same with Elastic IP as with a regular EC2 IP. Your existing SSL setup (Step 7) will work perfectly.

---

### 8.3 Point Domain to Elastic IP

**Option A: Using Your Domain Registrar's DNS (Simple)**
1. Go to your domain registrar (GoDaddy, Namecheap, Google Domains, etc.)
2. Navigate to **DNS Management** or **DNS Settings**
3. Add/Edit DNS A Record:
   - **Type:** A
   - **Name:** @ (or blank for root domain)
   - **Value:** Your Elastic IP address (e.g., `54.123.45.67`)
   - **TTL:** 3600 (or 600 for faster updates)

4. Add CNAME for www:
   - **Type:** CNAME
   - **Name:** www
   - **Value:** your-domain.com (or use the same Elastic IP)
   - **TTL:** 3600

**Option B: Using Cloudflare DNS (Recommended - FREE + Better Performance)**
1. Sign up for free account at [cloudflare.com](https://cloudflare.com)
2. Add your domain to Cloudflare
3. Update nameservers at your registrar (Cloudflare will provide them)
4. In Cloudflare DNS settings, add:
   - **Type:** A
   - **Name:** @
   - **IPv4 address:** Your Elastic IP
   - **Proxy status:** DNS only (gray cloud) - Important for API servers
   - **TTL:** Auto

5. Add CNAME for www:
   - **Type:** CNAME
   - **Name:** www
   - **Target:** your-domain.com
   - **Proxy status:** DNS only
   - **TTL:** Auto

**Why Cloudflare?**
- ✅ FREE DNS management
- ✅ FREE DDoS protection
- ✅ Faster DNS resolution globally
- ✅ Better reliability
- ✅ Full DNS control

**Note:** For API servers, use "DNS only" (gray cloud) mode, not "Proxied" (orange cloud), as proxying can interfere with WebSocket connections and direct IP access.

**Alternative:** If you don't use Elastic IP, use your EC2 Public IP (but it will change if instance restarts)

---

### 8.4 Wait for DNS Propagation
- Usually takes **5-60 minutes** (can take up to 48 hours)
- Check propagation with:
  ```bash
  nslookup your-domain.com
  # or
  dig your-domain.com
  ```
- Test from multiple locations: [whatsmydns.net](https://www.whatsmydns.net/)

---

### 8.5 Verify Domain Setup

Once DNS propagates:
```bash
# Test HTTP connection
curl -I http://your-domain.com

# Test HTTPS (after SSL setup)
curl -I https://your-domain.com
```

---

## Step 9: Set Up File Storage (S3 - Optional)

### 9.1 Create S3 Bucket
```bash
aws s3 mb s3://ar-13-uploads --region us-east-1
```

### 9.2 Configure Bucket Policy
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::ar-13-uploads/*"
    }
  ]
}
```

### 9.3 Update Application
Modify your file upload handler to use S3 instead of local storage.

---

## Step 10: Firewall Configuration

### 10.1 Configure UFW (Uncomplicated Firewall)
```bash
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
sudo ufw status
```

---

## Step 11: Monitoring and Maintenance

### 11.1 Set Up CloudWatch Monitoring
1. Go to EC2 → Instances → Select instance
2. Monitoring tab → Enable detailed monitoring (optional, costs extra)

### 11.2 Set Up CloudWatch Alarms
- CPU utilization > 80%
- Memory usage > 80%
- Disk usage > 80%

### 11.3 Regular Maintenance Commands
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Restart application
sudo systemctl restart ar13-backend

# Check application status
sudo systemctl status ar13-backend

# View logs
sudo journalctl -u ar13-backend -f

# Check disk space
df -h

# Check memory
free -h
```

---

## Troubleshooting

### Application Not Starting
```bash
# Check service status
sudo systemctl status ar13-backend

# Check logs
sudo journalctl -u ar13-backend -n 50

# Check if port is in use
sudo netstat -tlnp | grep 3000
```

### Nginx Errors
```bash
# Test Nginx configuration
sudo nginx -t

# Check Nginx logs
sudo tail -f /var/log/nginx/error.log
```

### SSL Certificate Issues
```bash
# Renew certificate manually
sudo certbot renew

# Check certificate status
sudo certbot certificates
```

### Database Connection Issues
```bash
# Test AWS credentials
aws dynamodb list-tables --region us-east-1

# Check IAM role permissions
aws sts get-caller-identity
```

---

## Security Best Practices

1. **Restrict SSH Access:**
   - Only allow your IP in security group
   - Use SSH keys, not passwords

2. **Keep System Updated:**
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

3. **Use Strong Passwords:**
   - Strong JWT_SECRET
   - Strong database passwords

4. **Enable AWS CloudTrail:**
   - Monitor API calls
   - Detect unauthorized access

5. **Regular Backups:**
   - Backup database regularly
   - Backup application code

---

## Cost Optimization

### Free Tier Usage:
- ✅ EC2 t2.micro: 750 hours/month (12 months)
- ✅ EBS Storage: 30 GB (12 months)
- ✅ Data Transfer: 100 GB out (12 months)
- ✅ DynamoDB: Forever free (25 GB, 25 RCU/WCU)
- ✅ S3: 5 GB storage forever free
- ✅ Elastic IP: 1 per account (FREE when associated with running instance)
- ❌ Route 53: **NO free tier** (~$0.50/month per hosted zone + query costs)

### Domain & DNS Costs:
- **Domain Registration:** ~$10-15/year (from registrar like GoDaddy, Namecheap)
- **DNS Management:** FREE (use registrar's DNS, not Route 53)
- **Elastic IP:** FREE (1 per account when associated with running instance)

### After 12 Months:
- EC2 t2.micro: ~$8-10/month
- EBS 20 GB: ~$2/month
- **Total: ~$10-12/month** (excluding domain registration)

---

## Next Steps

1. ✅ Test your API endpoints
2. ✅ Test WebSocket connection
3. ✅ Set up monitoring alerts
4. ✅ Configure backups
5. ✅ Set up CI/CD (optional)

---

## Quick Reference Commands

```bash
# Restart application
sudo systemctl restart ar13-backend

# View logs
sudo journalctl -u ar13-backend -f

# Check status
sudo systemctl status ar13-backend

# Restart Nginx
sudo systemctl restart nginx

# Test Nginx config
sudo nginx -t

# Renew SSL
sudo certbot renew

# Update application
cd ~/ar-13-go-backend
git pull
go build -o server cmd/server/main.go
sudo systemctl restart ar13-backend
```

---

## Support

If you encounter issues:
1. Check application logs: `sudo journalctl -u ar13-backend -f`
2. Check Nginx logs: `sudo tail -f /var/log/nginx/error.log`
3. Verify environment variables are set correctly
4. Check AWS IAM permissions
5. Verify security group rules

