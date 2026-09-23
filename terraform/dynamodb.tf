# DynamoDB is a different shape of AWS resource than RDS: there's no VPC,
# subnet group, or security group to wire up, because DynamoDB is a
# managed public AWS service reached over its API endpoint, not something
# that lives inside your VPC. Access control is entirely IAM-based instead
# of network-based — which is exactly what the IRSA setup below sets up.

resource "aws_dynamodb_table" "telemetry" {
  name         = "telemetry_readings"
  billing_mode = "PAY_PER_REQUEST" # on-demand pricing -- no capacity planning needed for a portfolio-scale table

  hash_key  = "satellite_id"
  range_key = "timestamp_ms"

  attribute {
    name = "satellite_id"
    type = "S"
  }

  attribute {
    name = "timestamp_ms"
    type = "N"
  }

  tags = {
    Project = var.project_name
  }
}

# --- IRSA: letting the grpc-server pod assume an IAM role, scoped to just
# this table, instead of using the EKS node's own (much broader) IAM role.
#
# The EKS module provisions an OIDC identity provider for the cluster by
# default. That's what lets a Kubernetes ServiceAccount "become" an AWS
# IAM role: the ServiceAccount gets annotated with a role ARN, EKS injects
# a short-lived web identity token into the pod, and AWS trusts that token
# because it was signed by an OIDC provider AWS itself was told to trust.
# This is the standard pattern for giving one specific workload narrow AWS
# permissions, instead of every pod on a node inheriting the node's IAM role.

data "aws_iam_policy_document" "grpc_server_assume_role" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    effect  = "Allow"

    principals {
      type        = "Federated"
      identifiers = [module.eks.oidc_provider_arn]
    }

    condition {
      test     = "StringEquals"
      variable = "${module.eks.oidc_provider}:sub"
      values   = ["system:serviceaccount:satellite-tracker:grpc-server"]
    }

    condition {
      test     = "StringEquals"
      variable = "${module.eks.oidc_provider}:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "grpc_server" {
  name               = "${var.project_name}-grpc-server"
  assume_role_policy = data.aws_iam_policy_document.grpc_server_assume_role.json

  tags = {
    Project = var.project_name
  }
}

# Scoped to exactly the actions and table this service needs -- not
# AmazonDynamoDBFullAccess, and not access to any other table.
data "aws_iam_policy_document" "grpc_server_dynamodb" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:PutItem",
      "dynamodb:Query",
      "dynamodb:DescribeTable",
    ]
    resources = [aws_dynamodb_table.telemetry.arn]
  }
}

resource "aws_iam_role_policy" "grpc_server_dynamodb" {
  name   = "dynamodb-telemetry-access"
  role   = aws_iam_role.grpc_server.id
  policy = data.aws_iam_policy_document.grpc_server_dynamodb.json
}
