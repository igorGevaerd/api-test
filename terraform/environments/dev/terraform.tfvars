environment = "dev"
aws_region  = "us-east-1"

app_name       = "api-test"
container_port = 8080

# Container configuration
container_cpu    = 256
container_memory = 512
desired_count    = 1

# Autoscaling
enable_autoscaling      = false
autoscaling_min_capacity = 1
autoscaling_max_capacity = 2

# Database
rds_instance_class  = "db.t3.micro"
rds_allocated_storage = 20

# ElastiCache
redis_node_type        = "cache.t3.micro"
redis_num_cache_nodes  = 1

# Logs
log_retention_days = 7

# Health checks
health_check_path                = "/health"
health_check_interval            = 30
health_check_timeout             = 5
health_check_healthy_threshold   = 2
health_check_unhealthy_threshold = 3

# Network
vpc_cidr               = "10.0.0.0/16"
public_subnet_cidrs    = ["10.0.1.0/24", "10.0.2.0/24"]
private_subnet_cidrs   = ["10.0.10.0/24", "10.0.11.0/24"]
enable_nat_gateway     = true

tags = {
  Environment = "dev"
  Terraform   = "true"
}
