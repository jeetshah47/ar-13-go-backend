# AR-13 Backend EC2 Instance Launcher
# Free Tier Configuration for 10 Users
# Usage: .\launch-ec2-instance.ps1 -Region "us-east-1" -KeyName "my-key" -CreateSecurityGroup

param(
    [Parameter(Mandatory=$false)]
    [string]$Region = "ap-south-1",
    
    [Parameter(Mandatory=$false)]
    [string]$KeyName = "",
    
    [Parameter(Mandatory=$false)]
    [string]$SecurityGroupId = "",
    
    [Parameter(Mandatory=$false)]
    [string]$SubnetId = "",
    
    [Parameter(Mandatory=$false)]
    [switch]$CreateSecurityGroup = $false,
    
    [Parameter(Mandatory=$false)]
    [string]$VpcId = ""
)

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Error { Write-Host $args -ForegroundColor Red }
function Write-Info { Write-Host $args -ForegroundColor Cyan }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }

Write-Info "=========================================="
Write-Info "AR-13 Backend EC2 Instance Launcher"
Write-Info "Free Tier Configuration (t3.micro)"
Write-Info "=========================================="
Write-Host ""

# Set region
Write-Info "Setting AWS region to: $Region"
aws configure set region $Region
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to set region. Please check AWS CLI installation."
    exit 1
}

# Get latest Amazon Linux 2023 AMI ID
Write-Info "Fetching latest Amazon Linux 2023 AMI for region: $Region..."
$amiId = aws ssm get-parameters `
    --names "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-6.1-x86_64" `
    --region $Region `
    --query "Parameters[0].Value" `
    --output text

if (-not $amiId -or $amiId -eq "None") {
    Write-Error "Failed to retrieve AMI ID. Trying alternative method..."
    # Try alternative SSM parameter
    $amiId = aws ssm get-parameters `
        --names "/aws/service/ami-amazon-linux-latest/al2023-ami-x86_64" `
        --region $Region `
        --query "Parameters[0].Value" `
        --output text
}

if (-not $amiId -or $amiId -eq "None") {
    Write-Error "Could not find AMI ID. Please check your region and AWS CLI configuration."
    exit 1
}

Write-Success "Using AMI: $amiId"
Write-Host ""

# Get or validate Key Pair
if (-not $KeyName) {
    Write-Info "Available key pairs:"
    aws ec2 describe-key-pairs --query "KeyPairs[*].KeyName" --output table
    Write-Host ""
    $KeyName = Read-Host "Enter key pair name"
    if (-not $KeyName) {
        Write-Error "Key pair name is required"
        exit 1
    }
}

# Validate key pair exists
$keyExists = aws ec2 describe-key-pairs --key-names $KeyName --query "KeyPairs[0].KeyName" --output text 2>$null
if (-not $keyExists) {
    Write-Error "Key pair '$KeyName' does not exist. Please create it first or use an existing key."
    exit 1
}
Write-Success "Using key pair: $KeyName"
Write-Host ""

# Get VPC ID if not provided
if (-not $VpcId) {
    Write-Info "Available VPCs:"
    aws ec2 describe-vpcs --query "Vpcs[*].[VpcId,IsDefault,CidrBlock]" --output table
    Write-Host ""
    $defaultVpc = aws ec2 describe-vpcs --filters "Name=isDefault,Values=true" --query "Vpcs[0].VpcId" --output text
    if ($defaultVpc) {
        $VpcId = $defaultVpc
        Write-Info "Using default VPC: $VpcId"
    } else {
        $VpcId = Read-Host "Enter VPC ID"
    }
}

# Get or create Security Group
if ($CreateSecurityGroup) {
    Write-Info "Creating security group..."
    $sgName = "ar-13-backend-sg"
    $sgDescription = "Security group for AR-13 backend - allows SSH and API access"
    
    # Check if security group already exists
    $existingSg = aws ec2 describe-security-groups `
        --filters "Name=group-name,Values=$sgName" "Name=vpc-id,Values=$VpcId" `
        --query "SecurityGroups[0].GroupId" `
        --output text 2>$null
    
    if ($existingSg) {
        Write-Warning "Security group '$sgName' already exists: $existingSg"
        $SecurityGroupId = $existingSg
    } else {
        # Get your public IP for SSH access
        $myIp = (Invoke-WebRequest -Uri "https://api.ipify.org" -UseBasicParsing).Content
        Write-Info "Your public IP detected: $myIp"
        
        $SecurityGroupId = aws ec2 create-security-group `
            --group-name $sgName `
            --description $sgDescription `
            --vpc-id $VpcId `
            --query "GroupId" `
            --output text
        
        if (-not $SecurityGroupId) {
            Write-Error "Failed to create security group"
            exit 1
        }
        
        Write-Success "Created security group: $SecurityGroupId"
        
        # Add SSH rule (from your IP)
        Write-Info "Adding SSH rule (port 22) from your IP: $myIp/32"
        aws ec2 authorize-security-group-ingress `
            --group-id $SecurityGroupId `
            --protocol tcp `
            --port 22 `
            --cidr "$myIp/32" `
            --output text | Out-Null
        
        # Add API rule (port 3000) - can be restricted later
        Write-Info "Adding API rule (port 3000) from anywhere (0.0.0.0/0)"
        Write-Warning "Consider restricting this to specific IPs in production!"
        aws ec2 authorize-security-group-ingress `
            --group-id $SecurityGroupId `
            --protocol tcp `
            --port 3000 `
            --cidr "0.0.0.0/0" `
            --output text | Out-Null
        
        Write-Success "Security group rules added"
    }
} elseif (-not $SecurityGroupId) {
    Write-Info "Available security groups:"
    aws ec2 describe-security-groups --filters "Name=vpc-id,Values=$VpcId" --query "SecurityGroups[*].[GroupId,GroupName,Description]" --output table
    Write-Host ""
    $SecurityGroupId = Read-Host "Enter security group ID (or use -CreateSecurityGroup to create one)"
    if (-not $SecurityGroupId) {
        Write-Error "Security group ID is required"
        exit 1
    }
}

Write-Success "Using security group: $SecurityGroupId"
Write-Host ""

# Get Subnet ID if not provided
if (-not $SubnetId) {
    Write-Info "Available subnets in VPC $VpcId :"
    aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VpcId" --query "Subnets[*].[SubnetId,AvailabilityZone,CidrBlock,MapPublicIpOnLaunch]" --output table
    Write-Host ""
    
    # Try to find a public subnet
    $publicSubnet = aws ec2 describe-subnets `
        --filters "Name=vpc-id,Values=$VpcId" "Name=map-public-ip-on-launch,Values=true" `
        --query "Subnets[0].SubnetId" `
        --output text
    
    if ($publicSubnet) {
        $SubnetId = $publicSubnet
        Write-Info "Using public subnet: $SubnetId"
    } else {
        $SubnetId = Read-Host "Enter subnet ID"
    }
}

if (-not $SubnetId) {
    Write-Error "Subnet ID is required"
    exit 1
}

Write-Success "Using subnet: $SubnetId"
Write-Host ""

# Create user data script
$userDataScript = @"
#!/bin/bash
# Update system
yum update -y

# Install Docker
yum install -y docker
systemctl start docker
systemctl enable docker
usermod -aG docker ec2-user

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-`$(uname -s)-`$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Create directories
mkdir -p /home/ec2-user/upload
mkdir -p /home/ec2-user/logs
chown -R ec2-user:ec2-user /home/ec2-user

# Install basic tools
yum install -y curl wget

# Configure swap (critical for 1GB RAM on t3.micro)
dd if=/dev/zero of=/swapfile bs=1M count=1024
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab

# Optimize for low memory
echo 'vm.swappiness=10' >> /etc/sysctl.conf
sysctl -p

# Log completion
echo "User data script completed at `$(date)" >> /var/log/user-data.log
"@

# Save user data to temp file
$userDataFile = [System.IO.Path]::GetTempFileName()
$userDataScript | Out-File -FilePath $userDataFile -Encoding ASCII -NoNewline

# Base64 encode user data
$userDataBase64 = [Convert]::ToBase64String([System.Text.Encoding]::ASCII.GetBytes($userDataScript))

Write-Info "Launching EC2 instance..."
Write-Info "Instance Type: t3.micro (Free Tier)"
Write-Info "Storage: 20 GB gp3 (Free Tier)"
Write-Info "CPU Credits: Unlimited"
Write-Host ""

# Launch instance
$launchResult = aws ec2 run-instances `
    --image-id $amiId `
    --instance-type t3.micro `
    --key-name $KeyName `
    --security-group-ids $SecurityGroupId `
    --subnet-id $SubnetId `
    --block-device-mappings "[{`"DeviceName`":`"/dev/xvda`",`"Ebs`":{`"VolumeSize`":20,`"VolumeType`":`"gp3`",`"DeleteOnTermination`":true}}]" `
    --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=ar-13-backend-free-tier},{Key=Project,Value=ar-13},{Key=Environment,Value=production}]" `
    --credit-specification "CpuCredits=unlimited" `
    --user-data $userDataBase64 `
    --region $Region `
    --output json

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to launch instance"
    Remove-Item $userDataFile -ErrorAction SilentlyContinue
    exit 1
}

# Parse instance ID
$instanceId = ($launchResult | ConvertFrom-Json).Instances[0].InstanceId
$privateIp = ($launchResult | ConvertFrom-Json).Instances[0].PrivateIpAddress

Write-Host ""
Write-Success "=========================================="
Write-Success "Instance Launched Successfully!"
Write-Success "=========================================="
Write-Host ""
Write-Info "Instance ID: $instanceId"
Write-Info "Private IP: $privateIp"
Write-Host ""

# Wait for instance to be running
Write-Info "Waiting for instance to be in 'running' state..."
$maxAttempts = 30
$attempt = 0
do {
    Start-Sleep -Seconds 5
    $state = aws ec2 describe-instances `
        --instance-ids $instanceId `
        --query "Reservations[0].Instances[0].State.Name" `
        --output text
    $attempt++
    Write-Host "  Attempt $attempt/$maxAttempts - State: $state" -NoNewline
    Write-Host "`r" -NoNewline
} while ($state -ne "running" -and $attempt -lt $maxAttempts)

Write-Host ""
Write-Host ""

# Get public IP
$publicIp = aws ec2 describe-instances `
    --instance-ids $instanceId `
    --query "Reservations[0].Instances[0].PublicIpAddress" `
    --output text

if ($publicIp) {
    Write-Success "Public IP: $publicIp"
    Write-Host ""
    Write-Info "You can SSH into the instance with:"
    Write-Host "  ssh -i $KeyName.pem ec2-user@$publicIp" -ForegroundColor Yellow
} else {
    Write-Warning "Public IP not yet assigned. It may take a few minutes."
    Write-Info "Check instance status: aws ec2 describe-instances --instance-ids $instanceId"
}

Write-Host ""
Write-Info "Next Steps:"
Write-Host "  1. Wait 2-3 minutes for user data script to complete"
Write-Host "  2. SSH into the instance"
Write-Host "  3. Create .env file with your configuration"
Write-Host "  4. Deploy your Docker image or build from source"
Write-Host "  5. Start the application with docker-compose"
Write-Host ""

# Cleanup
Remove-Item $userDataFile -ErrorAction SilentlyContinue

Write-Success "Launch complete! Check CloudWatch for logs if needed."





