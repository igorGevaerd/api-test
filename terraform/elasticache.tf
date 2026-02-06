# ElastiCache Subnet Group
resource "aws_elasticache_subnet_group" "main" {
  name       = "${var.app_name}-elasticache-subnet-group"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name = "${var.app_name}-elasticache-subnet-group"
  }
}

# ElastiCache Redis Cluster
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "${var.app_name}-redis"
  engine               = "redis"
  node_type            = var.redis_node_type
  num_cache_nodes      = var.redis_num_cache_nodes
  parameter_group_name = aws_elasticache_parameter_group.redis.name
  engine_version       = "7.0"
  port                 = 6379

  subnet_group_name          = aws_elasticache_subnet_group.main.name
  security_group_ids         = [aws_security_group.elasticache.id]
  
  transit_encryption_enabled = true
  
  snapshot_retention_limit = var.environment == "prod" ? 5 : 0
  snapshot_window          = "03:00-05:00"
  maintenance_window       = "sun:05:00-sun:06:00"

  log_delivery_configuration {
    destination      = aws_cloudwatch_log_group.redis.name
    destination_type = "cloudwatch-logs"
    log_format       = "json"
    log_type         = "engine-log"
  }

  notification_topic_arn = aws_sns_topic.elasticache_notifications.arn

  tags = {
    Name = "${var.app_name}-redis"
  }
}

# ElastiCache Parameter Group
resource "aws_elasticache_parameter_group" "redis" {
  family = "redis7"
  name   = "${var.app_name}-redis-params"

  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lru"
  }

  parameter {
    name  = "timeout"
    value = "300"
  }

  tags = {
    Name = "${var.app_name}-redis-params"
  }
}

# SNS Topic for ElastiCache notifications
resource "aws_sns_topic" "elasticache_notifications" {
  name = "${var.app_name}-elasticache-notifications"

  tags = {
    Name = "${var.app_name}-elasticache-notifications"
  }
}

resource "aws_sns_topic_subscription" "elasticache_notifications" {
  topic_arn = aws_sns_topic.elasticache_notifications.arn
  protocol  = "email"
  endpoint  = var.notification_email
}
