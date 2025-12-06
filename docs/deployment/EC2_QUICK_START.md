# EC2 Quick Start - Minimal Setup

## Quick Deployment Checklist

### 1. Launch EC2 Instance (5 minutes)
- [ ] Ubuntu 22.04 LTS
- [ ] t2.micro instance
- [ ] Security group: SSH (22), HTTP (80), HTTPS (443)
- [ ] Download key pair (.pem file)

### 2. Connect to Instance (1 minute)
```bash
chmod 400 your-key.pem
ssh -i your-key.pem ubuntu@YOUR_EC2_IP
```

### 3. Install Dependencies (5 minutes)
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Nginx
sudo apt install -y nginx

# Install Certbot
sudo apt install -y certbot python3-certbot-nginx

# Install Redis (optional)
sudo apt install -y redis-server
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

### 4. Deploy Application (5 minutes)
```bash
# Clone repository
cd ~
git clone YOUR_REPO_URL
cd ar-13-go-backend

# Build application
go mod download
go build -o server cmd/server/main.go

# Create .env file
nano .env
# (Add your environment variables)
```

### 5. Create Systemd Service (2 minutes)
```bash
sudo nano /etc/systemd/system/ar13-backend.service
```

Paste:
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

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable ar13-backend
sudo systemctl start ar13-backend
sudo systemctl status ar13-backend
```

### 6. Configure Nginx (3 minutes)

**Note:** Nginx configuration is the same whether you use Elastic IP or regular EC2 IP. No changes needed!

```bash
sudo nano /etc/nginx/sites-available/ar13-backend
```

Paste:
```nginx
server {
    listen 80;
    server_name YOUR_DOMAIN.com;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

Enable and test:
```bash
sudo ln -s /etc/nginx/sites-available/ar13-backend /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 7. Set Up SSL (2 minutes)
```bash
sudo certbot --nginx -d YOUR_DOMAIN.com
```

### 8. Configure Domain DNS (5 minutes)

**Important:** AWS Route 53 does NOT have a free tier for domain services. Use your domain registrar's DNS instead.

**Note:** AWS automatically creates a Public DNS like `ec2-98-86-113-245.compute-1.amazonaws.com`, but it changes when the instance restarts. Use your own domain instead.

**Option A: Use Elastic IP (Recommended - FREE)**
1. In EC2 Console → Elastic IPs → Allocate Elastic IP address
2. Associate it with your EC2 instance
3. Go to your domain registrar (GoDaddy, Namecheap, etc.)
4. Add A record: `@` → Your Elastic IP
5. Add CNAME: `www` → `your-domain.com`

**Note:** Elastic IP fully supports HTTPS. SSL certificates work the same way with Elastic IP.

**Option B: Use EC2 Public IP (Not Recommended)**
- IP changes when instance stops/restarts
- Add A record pointing to current EC2 IP
- Wait for DNS propagation (5-60 minutes)

**Total Time: ~30 minutes**

---

## Environment Variables Template

Create `.env` file with:
```env
PORT=3000
NODE_ENV=production
AWS_REGION=us-east-1
JWT_SECRET=your-32-character-secret-key-here
JWT_EXPIRATION_HOURS=24
REFRESH_EXPIRATION_DAYS=30
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USER=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM=noreply@yourdomain.com
EMAIL_FROM_NAME=AR-13
FRONTEND_URL=https://your-frontend-domain.com
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

---

## Useful Commands

```bash
# Restart app
sudo systemctl restart ar13-backend

# View logs
sudo journalctl -u ar13-backend -f

# Update app
cd ~/ar-13-go-backend
git pull
go build -o server cmd/server/main.go
sudo systemctl restart ar13-backend

# Check status
sudo systemctl status ar13-backend
```

---

## Troubleshooting

**App not starting?**
```bash
sudo journalctl -u ar13-backend -n 50
```

**Port already in use?**
```bash
sudo netstat -tlnp | grep 3000
```

**Nginx not working?**
```bash
sudo nginx -t
sudo tail -f /var/log/nginx/error.log
```

