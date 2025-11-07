# Required Dependencies for Migration

## Install AWS SDK and Other Dependencies

Run these commands to install all required packages:

```bash
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/dynamodb
go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue
go get golang.org/x/crypto/bcrypt
go get github.com/google/uuid
```

Or install all at once:

```bash
go get github.com/aws/aws-sdk-go-v2/config \
        github.com/aws/aws-sdk-go-v2/service/dynamodb \
        github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue \
        golang.org/x/crypto/bcrypt \
        github.com/google/uuid
```

After installing, run:

```bash
go mod tidy
```

## Package Versions

The migration uses:
- `github.com/aws/aws-sdk-go-v2` - AWS SDK v2 for Go
- `golang.org/x/crypto` - For bcrypt password hashing
- `github.com/google/uuid` - For generating user IDs

## AWS Credentials Setup

For local development, configure AWS credentials:

**Option 1: AWS Credentials File**
```bash
aws configure
```

**Option 2: Environment Variables**
```env
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-1
```

**Option 3: IAM Role (for EC2)**
If running on EC2, use IAM role instead of credentials.

