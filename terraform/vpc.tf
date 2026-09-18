# The VPC is the private network everything else lives inside: EKS nodes,
# and the RDS database. We use the widely-used terraform-aws-modules/vpc
# module instead of hand-writing subnets, route tables, and gateways —
# this is standard practice, most teams don't write VPC wiring by hand.
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${var.project_name}-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["${var.aws_region}a", "${var.aws_region}b"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24"] # EKS nodes and RDS live here, not directly reachable from the internet
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24"] # only load balancers / NAT gateway live here

  enable_nat_gateway   = true
  single_nat_gateway   = true # one shared NAT instead of one per AZ — cheaper, fine for a dev project
  enable_dns_hostnames = true

  # Required tags so the EKS control plane and its load balancer controller
  # can auto-discover which subnets belong to this cluster.
  public_subnet_tags = {
    "kubernetes.io/role/elb"                     = "1"
    "kubernetes.io/cluster/${var.project_name}"  = "shared"
  }
  private_subnet_tags = {
    "kubernetes.io/role/internal-elb"            = "1"
    "kubernetes.io/cluster/${var.project_name}"  = "shared"
  }

  tags = {
    Project = var.project_name
  }
}
