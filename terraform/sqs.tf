locals {
  queues = {
    analysis = "${var.project_name}-analysis"
    snapshot = "${var.project_name}-snapshot"
    ai       = "${var.project_name}-ai-explain"
  }
}

# Dead-letter queues
resource "aws_sqs_queue" "dlq" {
  for_each                  = local.queues
  name                      = "${each.value}-dlq"
  message_retention_seconds = 1209600 # 14 days

  tags = { Name = "${each.value}-dlq" }
}

# Main queues
resource "aws_sqs_queue" "main" {
  for_each                   = local.queues
  name                       = each.value
  visibility_timeout_seconds = 300 # 5 min — enough for analysis jobs
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 20 # long polling

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq[each.key].arn
    maxReceiveCount     = 3
  })

  tags = { Name = each.value }
}
