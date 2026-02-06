# Terraform ECS Fargate Infrastructure - Quick Start Guide

## 📋 What's Been Created

A complete production-ready Terraform infrastructure for deploying your Go API application to AWS ECS Fargate with:

### Infrastructure Components
- ✅ **VPC with Multi-AZ Support**: 2 public + 2 private subnets
- ✅ **Application Load Balancer**: With health checks and sticky sessions
- ✅ **ECS Fargate Cluster**: Serverless container orchestration
- ✅ **RDS PostgreSQL**: Database with backup and encryption
- ✅ **ElastiCache Redis**: In-memory cache with persistence
- ✅ **ECR Repository**: Private Docker image storage
- ✅ **IAM Roles**: Fine-grained permissions for tasks
- ✅ **Security Groups**: Network isolation and access control
- ✅ **CloudWatch Logs**: Centralized logging
- ✅ **CloudWatch Alarms**: Monitoring and alerts
- ✅ **Secrets Manager**: Secure credential storage
- ✅ **Auto Scaling**: Based on CPU and memory metrics

## 🗂️ File Structure

```
terraform/
├── versions.tf                          # Terraform & provider versions
├── variables.tf                         # Variable definitions (80+ variables)
├── outputs.tf                          # Output values (20+ outputs)
├── vpc.tf                              # VPC, subnets, NAT, routing
├── security_groups.tf                  # 4 security groups (ALB, ECS, RDS, Redis)
├── iam.tf                              # Task execution and task roles
├── ecr.tf                              # ECR repository with lifecycle policy
├── rds.tf                              # RDS PostgreSQL with encryption
├── elasticache.tf                      # Redis cluster with persistence
├── alb.tf                              # ALB, target groups, listeners
├── ecs.tf                              # ECS cluster, task definition, service
├── cloudwatch.tf                       # Logs, alarms, SNS notifications
├── README.md                           # Detailed documentation
├── deploy.sh                           # Automated deployment script
├── .gitignore                          # Git ignore patterns
├── terraform.tfvars.example            # Example variables
├── environments/
│   ├── dev/
│   │   ├── backend.tf                  # S3 backend configuration
│   │   └── terraform.tfvars            # Dev environment variables
│   └── prod/
│       ├── backend.tf                  # S3 backend configuration
│       └── terraform.tfvars            # Prod environment variables
└── DEPLOYMENT_GUIDE.md                 # Step-by-step deployment guide
```

## 🚀 Quick Start (5 Minutes)

### 1. Prerequisites
```bash
# Verify required tools
terraform version  # Should be 1.0+
aws --version      # Any recent version
docker --version   # Any recent version

# Configure AWS credentials
aws configure
```

### 2. Set Environment Variables
```bash
# Set database password
export TF_VAR_rds_password="YourSecurePassword123!@#"

# Optional: Set AWS region (default: us-east-1)
export AWS_REGION="us-east-1"
```

### 3. Deploy Development Environment
```bash
cd terraform/environments/dev

# Initialize
terraform init

# Plan
terraform plan

# Apply (takes 10-15 minutes)
terraform apply

# Get outputs
terraform output
```

### 4. Build and Push Docker Image
```bash
# Get ECR URL
ECR_URL=$(cd terraform/environments/dev && terraform output -raw ecr_repository_url)

# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin $(echo $ECR_URL | cut -d'/' -f1)

# Build and push
docker build -f docker/Dockerfile -t api-test:latest .
docker tag api-test:latest $ECR_URL:latest
docker push $ECR_URL:latest
```

### 5. Test Your API
```bash
# Get ALB URL
ALB_URL=$(cd terraform/environments/dev && terraform output -raw alb_url)

# Test health endpoint
curl http://$ALB_URL/health

# List users
curl http://$ALB_URL/users

# Create user
curl -X POST http://$ALB_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com"}'
```

## 📊 Infrastructure Details

### Development Environment
| Component | Configuration |
|-----------|---|
| Container CPU | 256 units |
| Container Memory | 512 MB |
| Desired Tasks | 1 |
| Auto Scaling | Disabled |
| RDS Instance | db.t3.micro |
| RDS Storage | 20 GB |
| Redis Node Type | cache.t3.micro |
| Backup Retention | 7 days |

### Production Environment
| Component | Configuration |
|-----------|---|
| Container CPU | 1024 units |
| Container Memory | 2048 MB |
| Desired Tasks | 3 |
| Auto Scaling | Enabled (3-10 tasks) |
| RDS Instance | db.t3.small |
| RDS Storage | 100 GB |
| Redis Node Type | cache.t3.small |
| Multi-AZ | Enabled |
| Backup Retention | 30 days |

## 🔧 Key Features

### Networking
- ✅ VPC with CIDR 10.0.0.0/16
- ✅ Public subnets for ALB
- ✅ Private subnets for ECS, RDS, Redis
- ✅ NAT Gateways for outbound traffic
- ✅ Internet Gateway for public access

### Security
- ✅ 4 isolated security groups
- ✅ Least privilege IAM roles
- ✅ Encryption at rest (RDS, Redis, Logs)
- ✅ Secrets Manager for passwords
- ✅ KMS keys for encryption
- ✅ No public database access

### Resilience
- ✅ Multi-AZ deployment (prod)
- ✅ Auto Scaling based on metrics
- ✅ Health checks on ALB
- ✅ RDS backups and encryption
- ✅ Redis persistence (AOF)
- ✅ CloudWatch monitoring and alarms

### Observability
- ✅ CloudWatch Logs for all services
- ✅ CloudWatch Alarms for metrics
- ✅ SNS notifications for alerts
- ✅ Container Insights enabled
- ✅ RDS enhanced monitoring

## 📝 Important Configuration Steps

### 1. Update Email Notifications
```bash
# Edit terraform/elasticache.tf and terraform/cloudwatch.tf
# Replace "your-email@example.com" with your email
sed -i '' 's/your-email@example.com/your.email@company.com/g' \
  terraform/elasticache.tf \
  terraform/cloudwatch.tf
```

### 2. Configure Backend S3 Bucket
```bash
# Create S3 bucket
aws s3api create-bucket \
  --bucket api-test-terraform-state-dev-$(date +%s) \
  --region us-east-1

# Update backend.tf with bucket name
sed -i '' "s/YOUR-TERRAFORM-STATE-BUCKET-DEV/your-bucket-name/g" \
  terraform/environments/dev/backend.tf
```

### 3. Set Database Password
```bash
# Option 1: Environment variable (recommended)
export TF_VAR_rds_password="YourSecurePassword123!@#"

# Option 2: terraform.tfvars
echo 'rds_password = "YourSecurePassword123!@#"' > \
  terraform/environments/dev/terraform.tfvars.secret
```

### 4. Configure HTTPS (Production)
```bash
# Update domain in terraform/alb.tf
sed -i '' 's/api.example.com/api.your-domain.com/g' terraform/alb.tf

# Set ACM certificate validation in Route53
# (Manual step required)
```

## 🚦 Deployment Commands

### Using Manual Terraform
```bash
cd terraform/environments/dev

# Full workflow
terraform init
terraform plan -out=tfplan
terraform apply tfplan

# View outputs
terraform output

# Destroy
terraform destroy
```

### Using Automated Script
```bash
# Plan deployment
terraform/deploy.sh dev plan

# Apply deployment
terraform/deploy.sh dev apply

# Destroy infrastructure
terraform/deploy.sh dev destroy

# View outputs
terraform/deploy.sh dev output
```

## 📈 Cost Estimation

### Development Environment
- **ECS Fargate**: $15-20/month
- **RDS t3.micro**: $10-15/month
- **ElastiCache t3.micro**: $10-15/month
- **ALB**: $16/month
- **NAT Gateway**: $0 (no traffic)
- **Data Transfer**: $0-5/month
- **Total**: $50-80/month

### Production Environment
- **ECS Fargate**: $80-120/month
- **RDS t3.small**: $30-50/month
- **ElastiCache t3.small**: $20-30/month
- **ALB**: $16/month
- **NAT Gateways**: $32/month
- **Data Transfer**: $10-30/month
- **Total**: $200-400/month

**Tip**: Use Reserved Instances to save 30-50%

## 🔍 Troubleshooting Quick Reference

### Check Service Status
```bash
aws ecs describe-services \
  --cluster api-test-cluster \
  --services api-test-service \
  --region us-east-1

# View task logs
aws logs tail /ecs/api-test-task --follow
```

### Check Database Status
```bash
aws rds describe-db-instances \
  --db-instance-identifier api-test-postgres \
  --query 'DBInstances[0].DBInstanceStatus'
```

### Check Cache Status
```bash
aws elasticache describe-cache-clusters \
  --cache-cluster-id api-test-redis \
  --query 'CacheClusters[0].CacheClusterStatus'
```

### Common Issues
| Issue | Solution |
|-------|----------|
| Tasks won't start | Check ECR image exists, verify IAM permissions |
| ALB shows unhealthy | Check security groups, verify app health endpoint |
| Can't connect to DB | Verify RDS security group, check credentials |
| High costs | Use smaller instances, enable autoscaling, use spot |

## 📚 Complete Documentation Files

1. **[terraform/README.md](terraform/README.md)** - Comprehensive Terraform documentation
2. **[DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)** - Step-by-step deployment guide
3. **[.github/WORKFLOWS_SETUP.md](.github/WORKFLOWS_SETUP.md)** - GitHub Actions CI/CD setup

## 🎯 Next Steps

After successful deployment:

1. ✅ **Set up CI/CD**: Configure GitHub Actions for automated builds
2. ✅ **Configure Monitoring**: Set up Datadog or similar for advanced monitoring
3. ✅ **Plan Backups**: Configure automated RDS snapshots
4. ✅ **Document Runbooks**: Create runbooks for common operations
5. ✅ **Set up Disaster Recovery**: Plan for failover scenarios
6. ✅ **Implement API Gateway**: Add rate limiting and authentication
7. ✅ **Add WAF**: Implement AWS WAF for security

## 🆘 Need Help?

### Check Logs
```bash
# ECS task logs
aws logs tail /ecs/api-test-task --follow

# RDS logs
aws logs tail /rds/api-test --follow

# Redis logs
aws logs tail /redis/api-test --follow
```

### Verify Permissions
```bash
# Check current AWS credentials
aws sts get-caller-identity

# Verify IAM permissions
aws iam list-attached-user-policies --user-name your-username
```

### Contact Support
- AWS Support: https://console.aws.amazon.com/support
- Terraform Docs: https://www.terraform.io/docs
- Go Documentation: https://golang.org/doc

## 📞 Quick Reference

**Environment Variables**:
```bash
export TF_VAR_rds_password="YourPassword"
export AWS_REGION="us-east-1"
export AWS_PROFILE="default"
```

**Terraform State**:
```bash
terraform state list              # List all resources
terraform state show aws_lb.main  # Show specific resource
terraform state rm resource_type.name  # Remove from state
```

**AWS CLI Quick Commands**:
```bash
# Get ALB URL
aws elbv2 describe-load-balancers --query 'LoadBalancers[0].DNSName'

# Get RDS endpoint
aws rds describe-db-instances --query 'DBInstances[0].Endpoint.Address'

# Get Redis endpoint
aws elasticache describe-cache-clusters --query 'CacheClusters[0].CacheNodes[0].Address'
```

---

## Summary

You now have a **production-ready** Terraform infrastructure that:
- Deploys to **AWS ECS Fargate** (no server management)
- Scales **automatically** based on demand
- Stores data in **PostgreSQL** with backups
- Caches data in **Redis** with persistence
- Monitors with **CloudWatch** and alarms
- Secures everything with **encryption and IAM**
- Works in **dev and prod** environments

**Total deployment time**: 15-20 minutes
**Lines of Terraform code**: 1000+
**AWS Services used**: 12+

Ready to deploy? Follow the **Quick Start** section above!

---

**Last Updated**: February 6, 2026
**Version**: 1.0
**Status**: ✅ Production Ready
