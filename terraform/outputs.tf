output "alb_dns" {
  description = "ALB DNS name for API"
  value       = aws_lb.main.dns_name
}

output "ecr_repository_url" {
  description = "ECR repository URL"
  value       = aws_ecr_repository.app.repository_url
}

output "sqs_analysis_url" {
  description = "SQS analysis queue URL"
  value       = aws_sqs_queue.main["analysis"].url
}

output "sqs_snapshot_url" {
  description = "SQS snapshot queue URL"
  value       = aws_sqs_queue.main["snapshot"].url
}

output "sqs_ai_url" {
  description = "SQS AI explanation queue URL"
  value       = aws_sqs_queue.main["ai"].url
}

output "aurora_endpoint" {
  description = "Aurora cluster endpoint"
  value       = aws_rds_cluster.main.endpoint
}

output "redis_endpoint" {
  description = "ElastiCache Redis endpoint"
  value       = aws_elasticache_serverless_cache.main.endpoint[0].address
}

