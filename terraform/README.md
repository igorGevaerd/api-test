# Terraform Infrastructure for ECS Fargate Deployment

This directory contains Terraform configurations to deploy the Go API application to AWS ECS Fargate with PostgreSQL, Redis, and Application Load Balancer.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                           Internet                               │
└────────────────────────────┬────────────────────────────────────┘
                             │
                    ┌────────▼──────────┐
                    │    ALB (80/443)   │
                    │  Security Group   │
                    └────────┬──────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        │           ┌────────▼──────────┐        │
        │           │  Public Subnet 1  │        │
        │           │   (AZ: us-east-1a)│       │
        │           └────────┬──────────┘        │
        │                    │                    │
        │           ┌────────▼──────────┐        │
        │           │  Public Subnet 2  │        │
        │           │   (AZ: us-east-1b)│       │
        │           └───────────────────┘        │
        │                                        │
┌───────┴─────────────────────────────────────────┴────────┐
│                        VPC (10.0.0.0/16)                  │
│                                                            │
│  ┌────────────────────────────────────────────────────┐   │
│  │  Private Subnets                                   │   │
│  │  10.0.10.0/24, 10.0.11.0/24                       │   │
│  │  (AZ: us-east-1a, us-east-1b)                     │   │
│  │                                                    │   │
│  │  ┌──────────────────────┐                         │   │
│  │  │  ECS Fargate Tasks   │                         │   │
│  │  │  (Go API)            │                         │   │
│  │  │  CPU: 256-1024       │                         │   │
│  │  │  Memory: 512-2048 MB │                         │   │
│  │  │  Count: 1-3 (dev/prod)                         │   │
│  │  │  AutoScaling: 2-10   │                         │   │
│  │  └──────────────────────┘                         │   │
│  │                                                    │   │
│  │  ┌──────────────────────┐                         │   │
│  │  │ RDS PostgreSQL       │                         │   │
│  │  │ Instance: t3.micro   │                         │   │
│  │  │ Storage: 20-100 GB   │                         │   │
│  │  │ Backup: 7-30 days    │                         │   │
│  │  │ Multi-AZ: (prod)     │                         │   │
│  │  └──────────────────────┘                         │   │
│  │                                                    │   │
│  │  ┌──────────────────────┐                         │   │
│  │  │ ElastiCache Redis    │                         │   │
│  │  │ Node: cache.t3.micro │                         │   │
│  │  │ AOF Persistence      │                         │   │
│  │  │ Encryption: At rest  │                         │   │
│  │  └──────────────────────┘                         │   │
│  │                                                    │   │
│  │  ┌──────────────────────┐                         │   │
│  │  │ NAT Gateways         │                         │   │
│  │  │ (Each AZ)            │                         │   │
│  │  └──────────────────────┘                         │   │
│  └────────────────────────────────────────────────────┘   │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
terraform/
├── versions.tf                 # Terraform and provider versions
├── variables.tf                # Variable definitions
├── outputs.tf                  # Output values
├── vpc.tf                      # VPC, subnets, NAT, route tables
├── security_groups.tf          # Security groups for ALB, ECS, RDS, Redis
├── iam.tf                      # IAM roles and policies
├── ecr.tf                      # ECR repository
├── rds.tf                      # RDS PostgreSQL database
├── elasticache.tf              # ElastiCache Redis cluster
├── alb.tf                      # Application Load Balancer
├── ecs.tf                      # ECS cluster, task definition, service
├── cloudwatch.tf               # CloudWatch logs and alarms
├── README.md                   # This file
├── environments/
│   ├── dev/
│   │   ├── terraform.tfvars    # Dev environment variables
│   │   └── backend.tf          # Dev backend configuration
│   └── prod/
│       ├── terraform.tfvars    # Prod environment variables
│       └── backend.tf          # Prod backend configuration
└── .gitignore                  # Terraform gitignore
```

## Prerequisites

1. **AWS Account**: Active AWS account with appropriate permissions
2. **Terraform**: Version 1.0 or higher
3. **AWS CLI**: Configured with appropriate credentials
4. **Docker**: For building and pushing images to ECR
5. **Go**: 1.21+ for building the application

## Setup Instructions

### 1. Create S3 Backend for State Management

```bash
# Create S3 bucket for dev state
aws s3api create-bucket \
  --bucket YOUR-TERRAFORM-STATE-BUCKET-DEV \
  --region us-east-1

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket YOUR-TERRAFORM-STATE-BUCKET-DEV \
  --versioning-configuration Status=Enabled

# Enable server-side encryption
aws s3api put-bucket-encryption \
  --bucket YOUR-TERRAFORM-STATE-BUCKET-DEV \
  --server-side-encryption-configuration '{
    "Rules": [
      {
        "ApplyServerSideEncryptionByDefault": {
          "SSEAlgorithm": "AES256"
        }
      }
    ]
  }'

# Create DynamoDB table for state locks
aws dynamodb create-table \
  --table-name terraform-locks-dev \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --region us-east-1

# Repeat for prod with -prod suffix
```

### 2. Update Backend Configuration

Edit `terraform/environments/dev/backend.tf` and `terraform/environments/prod/backend.tf`:

```bash
# Replace YOUR-TERRAFORM-STATE-BUCKET-DEV and YOUR-TERRAFORM-STATE-BUCKET-PROD
sed -i '' 's/YOUR-TERRAFORM-STATE-BUCKET-DEV/your-actual-bucket-dev/g' terraform/environments/dev/backend.tf
sed -i '' 's/YOUR-TERRAFORM-STATE-BUCKET-PROD/your-actual-bucket-prod/g' terraform/environments/prod/backend.tf
```

### 3. Set Database Password

```bash
# Set as environment variable (recommended)
export TF_VAR_rds_password="YourSecurePassword123!"

# Or in terraform.tfvars
echo 'rds_password = "YourSecurePassword123!"' >> terraform/environments/dev/terraform.tfvars
```

### 4. Initialize Terraform

```bash
cd terraform/environments/dev

# Initialize Terraform
terraform init

# Plan infrastructure
terraform plan -out=tfplan

# Review the plan and apply
terraform apply tfplan
```

### 5. Build and Push Docker Image

```bash
# Get ECR login token
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

# Get ECR repository URL from terraform output
ECR_URL=$(terraform output -raw ecr_repository_url)

# Build Docker image
docker build -f ../../docker/Dockerfile -t api-test:latest ../../

# Tag image
docker tag api-test:latest $ECR_URL:latest
docker tag api-test:latest $ECR_URL:$(date +%Y%m%d-%H%M%S)

# Push to ECR
docker push $ECR_URL:latest
docker push $ECR_URL:$(date +%Y%m%d-%H%M%S)
```

### 6. Update ECS Task Definition

After pushing the image, update the ECS task definition:

```bash
# Get current task definition
aws ecs describe-task-definition \
  --task-definition api-test \
  --region us-east-1 \
  --query taskDefinition | jq '.containerDefinitions[0].image = "'$ECR_URL':latest"' > task-def.json

# Register new task definition
aws ecs register-task-definition \
  --region us-east-1 \
  --cli-input-json file://task-def.json

# Update service to use new task definition
aws ecs update-service \
  --cluster api-test-cluster \
  --service api-test-service \
  --task-definition api-test \
  --region us-east-1
```

### 7. Configure DNS (Production)

For HTTPS support in production:

1. Update domain in `alb.tf`: Change `api.example.com` to your domain
2. Create Route53 hosted zone
3. Add A record pointing to ALB DNS name

## Configuration

### Environment Variables

Key environment variables controlled by Terraform:

- `PORT`: Container port (default: 8080)
- `ENVIRONMENT`: dev, staging, or prod
- `DB_HOST`: RDS endpoint
- `DB_PORT`: 5432
- `DB_NAME`: PostgreSQL database name
- `DB_USER`: PostgreSQL username
- `DB_PASSWORD`: Stored in Secrets Manager
- `REDIS_HOST`: ElastiCache endpoint
- `REDIS_PORT`: 6379

### Scaling Configuration

**Development**:
- Container CPU: 256 units
- Container Memory: 512 MB
- Desired Tasks: 1
- Auto Scaling: Disabled

**Production**:
- Container CPU: 1024 units
- Container Memory: 2048 MB
- Desired Tasks: 3
- Auto Scaling: Enabled (min: 3, max: 10)
- Target CPU: 70%
- Target Memory: 80%

### Database Configuration

**Development**:
- Instance Class: db.t3.micro
- Storage: 20 GB
- Backup Retention: 7 days
- Multi-AZ: Disabled

**Production**:
- Instance Class: db.t3.small
- Storage: 100 GB
- Backup Retention: 30 days
- Multi-AZ: Enabled
- Encryption: At-rest with KMS

### Caching Configuration

**Redis Cluster**:
- Engine: Redis 7.0
- Node Type: cache.t3.micro (dev), cache.t3.small (prod)
- Port: 6379
- Encryption: At-rest
- Persistence: AOF enabled (prod)
- Auto Failover: Enabled (prod)

## Terraform Commands

### Planning and Applying

```bash
cd terraform/environments/dev

# Initialize
terraform init

# Format code
terraform fmt -recursive

# Validate configuration
terraform validate

# Plan changes
terraform plan

# Apply changes
terraform apply

# Destroy infrastructure
terraform destroy
```

### Managing State

```bash
# List resources
terraform state list

# Show specific resource
terraform state show aws_lb.main

# Pull remote state
terraform state pull > state.json

# Push local state
terraform state push state.json
```

### Outputs

```bash
# Get all outputs
terraform output

# Get specific output
terraform output alb_url

# Get JSON format
terraform output -json
```

## Monitoring and Logging

### CloudWatch Logs

Logs are automatically collected for:

- **ECS Tasks**: `/ecs/api-test-task`
- **Redis**: `/redis/api-test`
- **RDS**: `/rds/api-test`

View logs:

```bash
# View recent ECS logs
aws logs tail /ecs/api-test-task --follow

# View logs from specific time
aws logs get-log-events \
  --log-group-name /ecs/api-test-task \
  --log-stream-name ecs/api-test/task-id \
  --start-time $(date -d '1 hour ago' +%s)000
```

### CloudWatch Alarms

Alarms are configured for:

- ECS CPU Utilization > 80%
- ECS Memory Utilization > 80%
- ALB Unhealthy Hosts >= 1

View alarms:

```bash
aws cloudwatch describe-alarms \
  --query 'MetricAlarms[?contains(AlarmName, `api-test`)]'
```

### Container Insights

ECS Container Insights is enabled for detailed monitoring:

1. Go to CloudWatch → Container Insights
2. Select cluster: `api-test-cluster`
3. View performance metrics and resource utilization

## Security Best Practices

### Implemented

✅ **Network Isolation**
- Private subnets for ECS, RDS, Redis
- Public subnets for ALB only
- NAT Gateways for outbound traffic

✅ **Encryption**
- RDS: At-rest encryption with KMS
- ElastiCache: At-rest encryption
- CloudWatch Logs: Encrypted with KMS
- Secrets Manager for database password

✅ **Access Control**
- Security groups with minimal permissions
- IAM roles with least privilege
- Secrets Manager for sensitive data

✅ **Monitoring**
- CloudWatch alarms for anomalies
- Container Insights enabled
- RDS enhanced monitoring

### Additional Recommendations

1. **Enable MFA** for AWS console access
2. **Use SSM Session Manager** instead of SSH
3. **Enable VPC Flow Logs** for network monitoring
4. **Configure CloudTrail** for audit logging
5. **Set up AWS Config** for compliance checking
6. **Use WAF** on ALB for DDoS protection
7. **Implement backup plan** for RDS
8. **Set resource limits** in AWS Organizations

## Cost Optimization

### Development Environment

- Use smaller instance types (t3.micro)
- Single AZ deployment
- Disable Multi-AZ
- Short backup retention (7 days)
- No auto-scaling

Estimated Cost: $50-80/month

### Production Environment

- Larger instance types (t3.small/medium)
- Multi-AZ for HA
- Auto-scaling enabled
- Extended backups (30 days)
- SNS notifications

Estimated Cost: $200-400/month

### Cost Reduction Tips

1. **Use Reserved Instances**: Save 30-50% with 1-year commitment
2. **Use Spot Instances**: For non-critical tasks (50-70% discount)
3. **Right-size instances**: Monitor metrics and adjust
4. **Clean up resources**: Delete unused snapshots, volumes
5. **Use lifecycle policies**: Auto-delete old logs

## Troubleshooting

### ECS Service Won't Start

```bash
# Check service events
aws ecs describe-services \
  --cluster api-test-cluster \
  --services api-test-service \
  --region us-east-1 \
  --query 'services[0].events'

# Check task logs
aws logs tail /ecs/api-test-task --follow

# Check task definition
aws ecs describe-task-definition \
  --task-definition api-test \
  --region us-east-1
```

### Database Connection Issues

```bash
# Check RDS status
aws rds describe-db-instances \
  --db-instance-identifier api-test-postgres \
  --query 'DBInstances[0].DBInstanceStatus'

# Check security group
aws ec2 describe-security-groups \
  --group-ids sg-xxxxx \
  --query 'SecurityGroups[0].IpPermissions'
```

### Redis Connection Issues

```bash
# Check ElastiCache status
aws elasticache describe-cache-clusters \
  --cache-cluster-id api-test-redis \
  --query 'CacheClusters[0].CacheClusterStatus'

# Test connectivity
redis-cli -h redis-endpoint -p 6379 ping
```

### Terraform Errors

```bash
# Validate configuration
terraform validate

# Check plan without apply
terraform plan -no-color > plan.txt

# Enable debug logging
export TF_LOG=DEBUG
terraform plan
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy to ECS

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          role-to-assume: arn:aws:iam::ACCOUNT_ID:role/GitHubActionsRole
          aws-region: us-east-1

      - name: Build and push Docker image
        run: |
          aws ecr get-login-password | docker login --username AWS --password-stdin ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com
          docker build -t $ECR_URL:$GITHUB_SHA .
          docker push $ECR_URL:$GITHUB_SHA

      - name: Update ECS task definition
        run: |
          aws ecs update-service \
            --cluster api-test-cluster \
            --service api-test-service \
            --force-new-deployment \
            --region us-east-1
```

## Useful Commands

```bash
# Get all resource information
terraform state list

# See specific resource details
terraform console
> aws_lb.main.dns_name

# Import existing resources
terraform import aws_instance.example i-1234567890abcdef0

# Show dependencies
terraform graph | dot -Tsvg > graph.svg

# Format all files
terraform fmt -recursive

# Validate all modules
terraform validate

# Lock and unlock state
terraform apply -lock=true
terraform apply -lock=false
```

## Additional Resources

- [Terraform AWS Provider Documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [AWS ECS Best Practices](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/best_practices.html)
- [AWS RDS Best Practices](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_BestPractices.html)
- [AWS ElastiCache Best Practices](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/BestPractices.html)
- [Terraform Best Practices](https://www.terraform.io/docs/cloud/guides/recommended-practices)

## Support and Contributions

For issues or questions:

1. Check CloudWatch logs
2. Review Terraform outputs
3. Validate AWS permissions
4. Check AWS service quotas
5. Contact AWS support for infrastructure issues

## License

This Terraform configuration is part of the api-test project.
