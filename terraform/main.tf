terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # State is kept locally on your machine by default (a terraform.tfstate
  # file). Fine for a solo portfolio project. A real team would use a
  # remote backend (S3 + DynamoDB locking) instead, so state isn't sitting
  # on one person's laptop.
}

provider "aws" {
  region = var.aws_region
}
