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
