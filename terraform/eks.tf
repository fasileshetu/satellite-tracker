# Same idea as the VPC: the EKS control plane involves IAM roles, security
# groups, and API server wiring that nobody hand-writes anymore. This module
# is the de facto standard for provisioning EKS with Terraform.
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = var.project_name
  cluster_version = "1.33" # extended support, keeps things a bit more stable than always chasing the newest minor version

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets # nodes run in private subnets, not exposed directly

  cluster_endpoint_public_access = true # lets kubectl on your laptop reach the cluster; a stricter setup would use a VPN/bastion instead

  eks_managed_node_groups = {
    default = {
      instance_types = [var.cluster_node_instance_type]
      min_size       = 1
      max_size       = 2
      desired_size   = var.cluster_node_desired_size
    }
  }

  # Without this, nobody (not even the person who created the cluster) can
  # view Kubernetes-level resources through kubectl or the console — EKS
  # access is separate from AWS IAM access. This grants full cluster-admin
  # to whichever IAM identity is running Terraform.
  access_entries = {
    admin = {
      principal_arn = data.aws_caller_identity.current.arn
      policy_associations = {
        admin = {
          policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
          access_scope = {
            type = "cluster"
          }
        }
      }
    }
  }

  tags = {
    Project = var.project_name
  }
}

data "aws_caller_identity" "current" {}