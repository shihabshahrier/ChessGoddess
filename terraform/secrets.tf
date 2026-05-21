resource "aws_secretsmanager_secret" "app" {
  name                    = "${var.project_name}/app-secrets"
  recovery_window_in_days = 0
  tags                    = { Name = "${var.project_name}-secrets" }
}

resource "aws_secretsmanager_secret_version" "app" {
  secret_id = aws_secretsmanager_secret.app.id
  secret_string = jsonencode({
    JWT_SECRET           = var.jwt_secret
    GOOGLE_CLIENT_ID     = var.google_client_id
    GOOGLE_CLIENT_SECRET = var.google_client_secret
    OPENROUTER_KEY       = var.openrouter_key
    DB_PASSWORD          = var.db_password
  })
}
