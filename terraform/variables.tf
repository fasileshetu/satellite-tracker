variable "aws_region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-west-2"
}

variable "project_name" {
  description = "Used as a prefix/tag on every resource, so they're easy to find and to tear down"
  type        = string
  default     = "satellite-tracker"
}

variable "cluster_node_instance_type" {
  description = "EC2 instance type for the EKS worker node(s). Kept small deliberately."
  type        = string
  default     = "t3.micro" 
}

variable "cluster_node_desired_size" {
  description = "Number of worker nodes. 1 is enough to prove this works; a real deployment would use more for redundancy."
  type        = number
  default     = 2
}

variable "db_instance_class" {
  description = "RDS instance class. db.t3.micro is free-tier eligible on a new AWS account."
  type        = string
  default     = "db.t3.micro"
}

variable "db_name" {
  type    = string
  default = "satellite_tracker"
}

variable "db_username" {
  type    = string
  default = "postgres"
}

variable "db_password" {
  description = "RDS master password. Passed in via terraform.tfvars (gitignored) or TF_VAR_db_password env var, never committed."
  type        = string
  sensitive   = true
}

variable "github_repo" {
  description = "owner/repo on GitHub, used to scope the OIDC trust policy so only workflow runs in this exact repo can assume the CI/CD IAM role"
  type        = string
  default     = "fasileshetu/satellite-tracker"
}
