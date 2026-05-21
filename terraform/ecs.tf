resource "aws_ecs_cluster" "main" {
  name = var.project_name

  setting {
    name  = "containerInsights"
    value = "disabled"
  }

  tags = { Name = "${var.project_name}-cluster" }
}

resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name = aws_ecs_cluster.main.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }
}

# Backend security group — ALB → ECS
resource "aws_security_group" "backend" {
  name_prefix = "${var.project_name}-backend-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${var.project_name}-backend-sg" }

  lifecycle { create_before_destroy = true }
}

locals {
  db_url    = "postgres://${var.db_username}:${var.db_password}@${aws_rds_cluster.main.endpoint}:5432/${var.db_name}?sslmode=require"
  redis_url = "rediss://${aws_elasticache_serverless_cache.main.endpoint[0].address}:6379"

  common_env = [
    { name = "ENVIRONMENT", value = "production" },
    { name = "PORT", value = "8080" },
    { name = "DATABASE_URL", value = local.db_url },
    { name = "REDIS_URL", value = local.redis_url },
    { name = "QUEUE_PROVIDER", value = "sqs" },
    { name = "SQS_ANALYSIS_URL", value = aws_sqs_queue.main["analysis"].url },
    { name = "SQS_SNAPSHOT_URL", value = aws_sqs_queue.main["snapshot"].url },
    { name = "SQS_AI_EXPLAIN_URL", value = aws_sqs_queue.main["ai"].url },
    { name = "ALLOWED_ORIGINS", value = var.frontend_url },
    { name = "FRONTEND_URL", value = var.frontend_url },
    { name = "GOOGLE_REDIRECT_URL", value = var.google_redirect_url },
    { name = "STOCKFISH_PATH", value = "/usr/bin/stockfish" },
  ]

  common_secrets = [
    {
      name      = "JWT_SECRET"
      valueFrom = "${aws_secretsmanager_secret.app.arn}:JWT_SECRET::"
    },
    {
      name      = "GOOGLE_CLIENT_ID"
      valueFrom = "${aws_secretsmanager_secret.app.arn}:GOOGLE_CLIENT_ID::"
    },
    {
      name      = "GOOGLE_CLIENT_SECRET"
      valueFrom = "${aws_secretsmanager_secret.app.arn}:GOOGLE_CLIENT_SECRET::"
    },
    {
      name      = "OPENROUTER_KEY"
      valueFrom = "${aws_secretsmanager_secret.app.arn}:OPENROUTER_KEY::"
    },
  ]
}

# --- API task definition ---
resource "aws_ecs_task_definition" "api" {
  family                   = "${var.project_name}-api"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.api_cpu
  memory                   = var.api_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name      = "api"
    image     = "${aws_ecr_repository.app.repository_url}:latest"
    essential = true

    portMappings = [{
      containerPort = 8080
      protocol      = "tcp"
    }]

    environment = concat(local.common_env, [
      { name = "HTTP_ENABLED", value = "true" },
      { name = "WORKER_ENABLED", value = "false" },
    ])

    secrets = local.common_secrets

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.app.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "api"
      }
    }
  }])

  tags = { Name = "${var.project_name}-api-task" }
}

resource "aws_ecs_service" "api" {
  name            = "${var.project_name}-api"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = var.api_desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.public[*].id
    security_groups  = [aws_security_group.backend.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "api"
    container_port   = 8080
  }

  depends_on = [aws_lb_listener.http]

  tags = { Name = "${var.project_name}-api-service" }
}

# --- Worker task definition (Fargate Spot — 70% cheaper) ---
resource "aws_ecs_task_definition" "worker" {
  family                   = "${var.project_name}-worker"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.worker_cpu
  memory                   = var.worker_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name      = "worker"
    image     = "${aws_ecr_repository.app.repository_url}:latest"
    essential = true

    environment = concat(local.common_env, [
      { name = "HTTP_ENABLED", value = "false" },
      { name = "WORKER_ENABLED", value = "true" },
    ])

    secrets = local.common_secrets

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.app.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "worker"
      }
    }
  }])

  tags = { Name = "${var.project_name}-worker-task" }
}

resource "aws_ecs_service" "worker" {
  name            = "${var.project_name}-worker"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.worker.arn
  desired_count   = var.worker_desired_count

  capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }

  network_configuration {
    subnets          = aws_subnet.public[*].id
    security_groups  = [aws_security_group.backend.id]
    assign_public_ip = true
  }

  tags = { Name = "${var.project_name}-worker-service" }
}
