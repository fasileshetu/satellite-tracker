output "eks_cluster_name" {
  value = module.eks.cluster_name
}

output "configure_kubectl" {
  description = "Run this after apply to point kubectl at the new cluster"
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${module.eks.cluster_name}"
}

output "rds_endpoint" {
  description = "Postgres host:port, reachable only from inside the VPC (i.e. from pods running in EKS)"
  value       = aws_db_instance.this.endpoint
}

output "database_url" {
  description = "Full connection string to hand to the API via a Kubernetes Secret"
  value       = "postgres://${var.db_username}:${var.db_password}@${aws_db_instance.this.endpoint}/${var.db_name}?sslmode=require"
  sensitive   = true
}

output "telemetry_table_name" {
  value = aws_dynamodb_table.telemetry.name
}

output "grpc_server_role_arn" {
  description = "Paste this into k8s/aws/grpc-server-deployment.yaml's ServiceAccount annotation (eks.amazonaws.com/role-arn)"
  value       = aws_iam_role.grpc_server.arn
}

output "cognito_user_pool_id" {
  value = aws_cognito_user_pool.this.id
}

output "cognito_app_client_id" {
  value = aws_cognito_user_pool_client.this.id
}

output "cognito_issuer_url" {
  description = "The OIDC issuer -- the Go middleware fetches <this>/.well-known/openid-configuration and the JWKS it points to"
  value       = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.this.id}"
}

output "cognito_hosted_ui_url" {
  description = "Cognito's free hosted login page, for testing the real Authorization Code redirect flow"
  value       = "https://${aws_cognito_user_pool_domain.this.domain}.auth.${var.aws_region}.amazoncognito.com/login?client_id=${aws_cognito_user_pool_client.this.id}&response_type=code&scope=openid+email+profile&redirect_uri=http://localhost:3000/callback"
}
