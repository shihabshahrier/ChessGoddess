resource "aws_elasticache_serverless_cache" "main" {
  engine = "redis"
  name   = "${var.project_name}-redis"

  cache_usage_limits {
    data_storage {
      maximum = var.elasticache_max_memory_gb
      unit    = "GB"
    }
  }

  subnet_ids         = aws_subnet.private[*].id
  security_group_ids = [aws_security_group.elasticache.id]

  tags = { Name = "${var.project_name}-redis" }
}

resource "aws_security_group" "elasticache" {
  name_prefix = "${var.project_name}-redis-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.backend.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${var.project_name}-redis-sg" }

  lifecycle { create_before_destroy = true }
}
