# ECS Fargate Deployment Guide

Complete step-by-step guide to deploy your Go API application to AWS ECS Fargate using Terraform.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Step 1: Set Up AWS Account](#step-1-set-up-aws-account)
3. [Step 2: Prepare Terraform Backend](#step-2-prepare-terraform-backend)
4. [Step 3: Configure Terraform](#step-3-configure-terraform)
5. [Step 4: Deploy Infrastructure](#step-4-deploy-infrastructure)
6. [Step 5: Build and Push Docker Image](#step-5-build-and-push-docker-image)
7. [Step 6: Verify Deployment](#step-6-verify-deployment)
8. [Step 7: Monitor and Maintain](#step-7-monitor-and-maintain)

## Prerequisites

### Required Tools

```bash
# Check Terraform version (need 1.0+)
terraform version

# Check AWS CLI version
aws --version

# Check Docker version
docker --version

# Check Go version
go version

# Install AWS CLI if needed
curl "https://awscli.amazonaws.com/awscli-exe-darwin-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Install Terraform if needed
brew install terraform

# Install jq for JSON parsing
brew install jq
```

### AWS Account Setup

1. **Create AWS Account** at https://aws.amazon.com
2. **Create IAM User** with programmatic access:
   - Go to IAM Console
   - Create user with policies:
     - `AmazonEC2FullAccess`
     - `AmazonECS_FullAccess`
     - `AmazonRDSFullAccess`
     - `AmazonElastiCacheFullAccess`
     - `IAMFullAccess`
     - `AmazonVPCFullAccess`
     - `AWSCloudFormationFullAccess`
     - `CloudWatchFullAccess`
     - `AWSSecretsManagerFullAccess`
     - `AmazonEC2ContainerRegistryFullAccess`
3. **Save Access Key ID and Secret Access Key**
4. **Configure AWS CLI**:

```bash
aws configure
# Enter Access Key ID
# Enter Secret Access Key
# Enter region: us-east-1
# Enter output format: json
```

## Step 1: Set Up AWS Account

### 1.1 Verify AWS Credentials

```bash
# Test AWS CLI connectivity
aws sts get-caller-identity
# Should output your account information

# Verify permissions
aws ec2 describe-regions
aws iam list-users
```

### 1.2 Set Up S3 and DynamoDB for State Management

```bash
# Create S3 bucket for Terraform state (dev)
aws s3api create-bucket \
  --bucket api-test-terraform-state-dev-$(date +%s) \
  --region us-east-1

# Store bucket name
export TF_STATE_BUCKET_DEV="api-test-terraform-state-dev-1707216000"

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket $TF_STATE_BUCKET_DEV \
  --versioning-configuration Status=Enabled

# Enable encryption
aws s3api put-bucket-encryption \
  --bucket $TF_STATE_BUCKET_DEV \
  --server-side-encryption-configuration '{
    "Rules": [
      {
        "ApplyServerSideEncryptionByDefault": {
          "SSEAlgorithm": "AES256"
        }
      }
    ]
  }'

# Block public access
aws s3api put-public-access-block \
  --bucket $TF_STATE_BUCKET_DEV \
  --public-access-block-configuration \
  "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

# Create DynamoDB table for state locks
aws dynamodb create-table \
  --table-name terraform-locks-dev \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --region us-east-1

# Enable TTL
aws dynamodb update-time-to-live \
  --table-name terraform-locks-dev \
  --time-to-live-specification "Enabled=true, AttributeName=ExpirationTime" \
  --region us-east-1
```

### 1.3 Repeat for Production (if needed)

```bash
# Replace 'dev' with 'prod' in above commands
# Or use provided script
```

## Step 2: Prepare Terraform Backend

### 2.1 Update Backend Configuration

```bash
# Update dev backend
sed -i '' "s/YOUR-TERRAFORM-STATE-BUCKET-DEV/$TF_STATE_BUCKET_DEV/g" \
  terraform/environments/dev/backend.tf

# Update prod backend (if using prod)
sed -i '' "s/YOUR-TERRAFORM-STATE-BUCKET-PROD/$TF_STATE_BUCKET_PROD/g" \
  terraform/environments/prod/backend.tf

# Verify changes
cat terraform/environments/dev/backend.tf
```

### 2.2 Set RDS Password

```bash
# Option 1: Set environment variable (recommended)
export TF_VAR_rds_password="YourSecurePassword123!@#"

# Option 2: Create terraform.tfvars in environment folder
echo 'rds_password = "YourSecurePassword123!@#"' > terraform/environments/dev/terraform.tfvars.secret

# Make sure to add to .gitignore
echo "terraform.tfvars.secret" >> terraform/.gitignore
```

## Step 3: Configure Terraform

### 3.1 Review and Customize Variables

```bash
# Review dev variables
cat terraform/environments/dev/terraform.tfvars

# Review prod variables
cat terraform/environments/prod/terraform.tfvars

# Customize as needed:
# - container_cpu and container_memory
# - rds_instance_class
# - redis_node_type
# - autoscaling parameters
# - vpc_cidr and subnet CIDRs
```

### 3.2 Update Email Notifications

Edit `terraform/elasticache.tf` and `terraform/cloudwatch.tf`:

```bash
# Find and replace notification email
sed -i '' 's/your-email@example.com/your.email@company.com/g' \
  terraform/elasticache.tf \
  terraform/cloudwatch.tf

# Update domain for HTTPS (if using prod)
sed -i '' 's/api.example.com/api.your-domain.com/g' terraform/alb.tf
```

### 3.3 Initialize Terraform

```bash
cd terraform/environments/dev

# Initialize
terraform init

# Validate
terraform validate

# Format
terraform fmt -recursive
```

## Step 4: Deploy Infrastructure

### 4.1 Plan Deployment

```bash
cd terraform/environments/dev

# Create plan
terraform plan -out=tfplan

# Review the plan carefully!
# Look for any unexpected changes
```

### 4.2 Apply Configuration

```bash
# Apply the plan
terraform apply tfplan

# This will take 10-15 minutes

# Save outputs
terraform output > outputs.json
```

### 4.3 Review Outputs

```bash
# Get important URLs and endpoints
terraform output alb_url
terraform output ecr_repository_url
terraform output rds_address
terraform output redis_endpoint

# Save for later use
export ALB_URL=$(terraform output -raw alb_url)
export ECR_URL=$(terraform output -raw ecr_repository_url)
export RDS_ENDPOINT=$(terraform output -raw rds_address)
export REDIS_ENDPOINT=$(terraform output -raw redis_endpoint)

echo "ALB URL: $ALB_URL"
echo "ECR URL: $ECR_URL"
echo "RDS Endpoint: $RDS_ENDPOINT"
echo "Redis Endpoint: $REDIS_ENDPOINT"
```

## Step 5: Build and Push Docker Image

### 5.1 Prepare Docker Image

```bash
cd /path/to/api-test

# Set ECR repository URL
export ECR_URL=$(cd terraform/environments/dev && terraform output -raw ecr_repository_url)
export AWS_REGION="us-east-1"
export ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# Get ECR login
aws ecr get-login-password --region $AWS_REGION | \
  docker login --username AWS --password-stdin $ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com
```

### 5.2 Build Docker Image

```bash
# Build from Dockerfile
docker build -f docker/Dockerfile -t api-test:latest .

# Check image
docker images | grep api-test
```

### 5.3 Tag and Push Image

```bash
# Get timestamp for version
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
GIT_HASH=$(git rev-parse --short HEAD)

# Tag images
docker tag api-test:latest $ECR_URL:latest
docker tag api-test:latest $ECR_URL:$TIMESTAMP
docker tag api-test:latest $ECR_URL:$GIT_HASH

# Push to ECR
docker push $ECR_URL:latest
docker push $ECR_URL:$TIMESTAMP
docker push $ECR_URL:$GIT_HASH

# Verify in ECR
aws ecr describe-images --repository-name api-test --region $AWS_REGION
```

## Step 6: Verify Deployment

### 6.1 Check ECS Service Status

```bash
# Get service status
aws ecs describe-services \
  --cluster api-test-cluster \
  --services api-test-service \
  --region us-east-1 \
  --query 'services[0].{Status:status,DesiredCount:desiredCount,RunningCount:runningCount}'

# Watch service updates
aws ecs describe-services \
  --cluster api-test-cluster \
  --services api-test-service \
  --region us-east-1 \
  --query 'services[0].events[0:5]'
```

### 6.2 Check ALB Health

```bash
# Get target group health
aws elbv2 describe-target-health \
  --target-group-arn $(aws elbv2 describe-target-groups \
    --names api-test-tg --region us-east-1 \
    --query 'TargetGroups[0].TargetGroupArn' --output text) \
  --region us-east-1

# Should show targets as "healthy"
```

### 6.3 Test API Endpoint

```bash
# Get ALB DNS name
ALB_URL=$(cd terraform/environments/dev && terraform output -raw alb_url)

# Test health endpoint
curl http://$ALB_URL/health

# Should return: {"status":"ok"}

# List users
curl http://$ALB_URL/users

# Create user
curl -X POST http://$ALB_URL/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com"
  }'
```

### 6.4 Check Logs

```bash
# View recent logs
aws logs tail /ecs/api-test-task --follow --region us-east-1

# View logs from specific time
aws logs get-log-events \
  --log-group-name /ecs/api-test-task \
  --log-stream-name ecs/api-test/$(aws ecs list-tasks \
    --cluster api-test-cluster \
    --service-name api-test-service \
    --region us-east-1 \
    --query 'taskArns[0]' --output text | awk -F/ '{print $NF}') \
  --region us-east-1
```

## Step 7: Monitor and Maintain

### 7.1 Set Up Monitoring

```bash
# View CloudWatch alarms
aws cloudwatch describe-alarms \
  --query 'MetricAlarms[?contains(AlarmName, `api-test`)]' \
  --region us-east-1

# Subscribe to SNS notifications
# Check your email for SNS subscription confirmations
```

### 7.2 Database Management

```bash
# Connect to RDS database
PGPASSWORD=$TF_VAR_rds_password psql \
  -h $(terraform output -raw rds_address) \
  -U apiuser \
  -d apidb

# List tables
\dt

# Check user count
SELECT COUNT(*) FROM users;

# Exit
\q
```

### 7.3 Cache Management

```bash
# Install redis-cli if needed
brew install redis

# Connect to Redis
redis-cli -h $(terraform output -raw redis_endpoint) -p 6379

# Check keys
KEYS *

# Get connection info
INFO

# Exit
exit
```

### 7.4 Regular Maintenance Tasks

```bash
# Weekly: Check resource costs
aws ce get-cost-and-usage \
  --time-period Start=2024-01-01,End=2024-01-31 \
  --granularity MONTHLY \
  --metrics BlendedCost \
  --group-by Type=DIMENSION,Key=SERVICE

# Monthly: Review backups
aws rds describe-db-instances \
  --db-instance-identifier api-test-postgres \
  --query 'DBInstances[0].{BackupRetention:BackupRetentionPeriod,LatestSnapshot:LatestRestorableTime}'

# Review auto-scaling metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name CPUUtilization \
  --dimensions Name=ServiceName,Value=api-test-service \
  --start-time $(date -d '7 days ago' -I)T00:00:00Z \
  --end-time $(date -I)T00:00:00Z \
  --period 3600 \
  --statistics Average
```

## Troubleshooting

### Tasks Won't Start

```bash
# Check task definition
aws ecs describe-task-definition \
  --task-definition api-test \
  --region us-east-1

# Check for errors in logs
aws logs tail /ecs/api-test-task --follow

# Common issues:
# 1. Docker image not found - verify ECR push
# 2. Insufficient permissions - check IAM roles
# 3. Network issues - check security groups
```

### Database Connection Issues

```bash
# Check RDS status
aws rds describe-db-instances \
  --db-instance-identifier api-test-postgres \
  --query 'DBInstances[0].DBInstanceStatus'

# Check security group allows ECS tasks
aws ec2 describe-security-groups \
  --group-ids sg-xxxxx \
  --query 'SecurityGroups[0].IpPermissions'

# Test connectivity from EC2 bastion
# (You may need to create an EC2 instance in the VPC)
```

### Rollback Changes

```bash
# Revert to previous version
git checkout HEAD~1 -- terraform/

# Or restore from S3 state
aws s3 cp s3://$TF_STATE_BUCKET_DEV/api-test/dev/terraform.tfstate.backup .

# Or use Terraform state command
terraform state pull > backup.json
aws s3 cp backup.json s3://$TF_STATE_BUCKET_DEV/backups/
terraform state push backup.json
```

## Cleanup

### Destroy Development Environment

```bash
cd terraform/environments/dev

# Plan destruction
terraform plan -destroy

# Destroy resources
terraform destroy

# Verify all resources deleted
aws ec2 describe-instances --filters "Name=instance.state.name,Values=running" --region us-east-1
aws rds describe-db-instances --region us-east-1
aws elasticache describe-cache-clusters --region us-east-1
```

### Keep Production Safe

```bash
# For production, use:
# 1. Terraform state locking (automatic via DynamoDB)
# 2. Deletion protection on RDS
# 3. Require approval for terraform destroy
# 4. Keep state backups
```

## Next Steps

1. ✅ Infrastructure is deployed
2. ✅ Application is running
3. Now:
   - Set up CI/CD pipeline (GitHub Actions)
   - Configure auto-scaling policies
   - Set up monitoring alerts
   - Plan backup strategy
   - Document runbooks
   - Plan disaster recovery

## Useful Resources

- [AWS ECS Documentation](https://docs.aws.amazon.com/ecs/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest)
- [AWS Fargate Pricing](https://aws.amazon.com/fargate/pricing/)
- [Terraform Cloud](https://app.terraform.io/)

## Support

For issues:
1. Check CloudWatch logs: `/ecs/api-test-task`
2. Review Terraform state: `terraform state list`
3. Verify AWS permissions
4. Check service quotas: `aws service-quotas list-service-quotas`

---

**Last Updated**: February 6, 2026
**Version**: 1.0
