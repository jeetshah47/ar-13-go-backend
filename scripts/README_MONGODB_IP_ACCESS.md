# MongoDB Atlas IP Access Script

This script automatically adds your current IP address (or a specified IP) to your MongoDB Atlas IP access list.

## Prerequisites

1. **MongoDB Atlas Account**: You need an active MongoDB Atlas account
2. **API Keys**: You need to create API keys with appropriate permissions
3. **Project ID**: You need your MongoDB Atlas Project ID

## Required Credentials

### 1. MongoDB Atlas Public API Key

**How to get it:**
1. Log in to [MongoDB Atlas](https://cloud.mongodb.com)
2. Navigate to **Access Manager** → **API Keys**
3. Click **Create API Key**
4. Give it a name (e.g., "IP Access Script")
5. Select **Project Owner** or **Organization Owner** role
6. Copy the **Public Key** (starts with something like `abcdefgh`)

### 2. MongoDB Atlas Private API Key

**How to get it:**
1. After creating the API key, you'll see the **Private Key**
2. **IMPORTANT**: Copy this immediately - it's only shown once!
3. If you lose it, you'll need to delete and recreate the API key
4. The private key looks like: `a1b2c3d4-e5f6-7890-abcd-ef1234567890`

### 3. MongoDB Atlas Project ID

**How to get it:**
1. In MongoDB Atlas, go to **Project Settings** → **General**
2. Find the **Project ID** field
3. Copy the Project ID (looks like: `507f1f77bcf86cd799439011`)

## Usage

### Option 1: Using Environment Variables (Recommended)

Set the following environment variables:

```bash
# Windows PowerShell
$env:MONGODB_ATLAS_PUBLIC_KEY="your-public-key"
$env:MONGODB_ATLAS_PRIVATE_KEY="your-private-key"
$env:MONGODB_ATLAS_PROJECT_ID="your-project-id"

# Windows CMD
set MONGODB_ATLAS_PUBLIC_KEY=your-public-key
set MONGODB_ATLAS_PRIVATE_KEY=your-private-key
set MONGODB_ATLAS_PROJECT_ID=your-project-id

# Linux/Mac
export MONGODB_ATLAS_PUBLIC_KEY="your-public-key"
export MONGODB_ATLAS_PRIVATE_KEY="your-private-key"
export MONGODB_ATLAS_PROJECT_ID="your-project-id"
```

Then run:
```bash
go run scripts/add_mongodb_ip_access.go
```

### Option 2: Using Command-Line Flags

```bash
go run scripts/add_mongodb_ip_access.go \
  -public-key "your-public-key" \
  -private-key "your-private-key" \
  -project-id "your-project-id"
```

### Option 3: Using .env File

Create a `.env` file in the project root (make sure it's in `.gitignore`):

```env
MONGODB_ATLAS_PUBLIC_KEY=your-public-key
MONGODB_ATLAS_PRIVATE_KEY=your-private-key
MONGODB_ATLAS_PROJECT_ID=your-project-id
```

The script will automatically load it if you have `godotenv` configured.

## Advanced Usage

### Add a Specific IP Address

```bash
go run scripts/add_mongodb_ip_access.go -ip "192.168.1.100"
```

### Add a Comment/Description

```bash
go run scripts/add_mongodb_ip_access.go -comment "Development Server - Office"
```

### Combine Options

```bash
go run scripts/add_mongodb_ip_access.go \
  -ip "203.0.113.42" \
  -comment "Production Server" \
  -public-key "your-public-key" \
  -private-key "your-private-key" \
  -project-id "your-project-id"
```

## Features

- ✅ Automatically detects your current public IP address
- ✅ Checks if IP already exists before adding (prevents duplicates)
- ✅ Supports adding specific IP addresses
- ✅ Allows custom comments/descriptions
- ✅ Uses MongoDB Atlas Admin API
- ✅ Provides clear error messages

## Security Notes

⚠️ **Important Security Considerations:**

1. **Never commit API keys to version control**
   - Add `.env` to `.gitignore` if using environment variables
   - Never commit files containing API keys

2. **API Key Permissions**
   - Use the minimum required permissions (Project Owner is usually sufficient)
   - Consider creating a separate API key just for IP management

3. **Private Key Storage**
   - Store private keys securely (use a password manager)
   - Rotate keys periodically
   - Delete unused API keys

4. **IP Access List**
   - Regularly review and remove unused IP addresses
   - Use specific IPs instead of `0.0.0.0/0` (allow all) in production

## Troubleshooting

### Error: "API returned status 401"
- Check that your API keys are correct
- Ensure the API key hasn't been deleted or disabled
- Verify the public and private keys are in the correct order

### Error: "API returned status 403"
- Check that your API key has sufficient permissions (Project Owner or Organization Owner)
- Verify you're using the correct Project ID

### Error: "API returned status 400"
- The IP address format might be invalid
- Check that the IP address is a valid IPv4 address

### Error: "Failed to get current public IP"
- Check your internet connection
- The ipify.org service might be temporarily unavailable
- Try using the `-ip` flag to specify an IP manually

## Example Output

```
Detecting current public IP address...
Detected public IP address: 203.0.113.42
Checking if IP is already in access list...
Adding IP address 203.0.113.42 to MongoDB Atlas access list...
Successfully added IP address 203.0.113.42 to MongoDB Atlas access list!
```

## Related Documentation

- [MongoDB Atlas API Documentation](https://www.mongodb.com/docs/atlas/reference/api-resources/)
- [MongoDB Atlas IP Access List](https://www.mongodb.com/docs/atlas/security/ip-access-list/)

