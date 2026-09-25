# Lets GitHub Actions authenticate to AWS without any long-lived access
# keys sitting in repo secrets. GitHub's OIDC token service issues a
# short-lived, workflow-scoped identity token on every run; AWS trusts it
# because of the federated identity provider below, and hands back
# temporary credentials for exactly the one IAM role we define here.
resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  # GitHub's OIDC token-signing certificate thumbprint. AWS actually
  # verifies the certificate chain itself now and largely ignores this
  # value, but the provider resource still requires one.
  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1"]
}

# Scopes trust to this one repo -- any workflow run in
# fasileshetu/satellite-tracker (any branch, any PR) can assume this role,
# but no other GitHub repo, org, or user can. Tightening this further to a
# single branch (repo:fasileshetu/satellite-tracker:ref:refs/heads/main)
# would mean only pushes to main can deploy, which is what a real
# production pipeline would do -- left broader here so PR-triggered test
# runs still work without a second role.
data "aws_iam_policy_document" "github_actions_trust" {
  statement {
    # aws-actions/configure-aws-credentials@v4 tags the assumed-role
    # session with GitHub run context (repo, ref, actor, workflow, ...)
    # by default, which means the actual API call is
    # AssumeRoleWithWebIdentity *and* TagSession together. Without
    # sts:TagSession allowed here too, AWS rejects the whole request as
    # unauthorized -- which is exactly what happened on the first two
    # workflow runs, even though AssumeRoleWithWebIdentity alone was
    # correctly configured.
    actions = ["sts:AssumeRoleWithWebIdentity", "sts:TagSession"]
    effect  = "Allow"

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:${var.github_repo}:*"]
    }
  }
}

resource "aws_iam_role" "github_actions" {
  name               = "${var.project_name}-github-actions"
  assume_role_policy = data.aws_iam_policy_document.github_actions_trust.json
}

# Deliberately scoped to just what the pipeline does -- push images to
# these two ECR repos and read enough EKS/STS info to run kubectl.
# Nothing here grants it the ability to touch other AWS resources, unlike
# the cluster-admin access_entry in eks.tf that's meant for a human
# running Terraform by hand.
data "aws_iam_policy_document" "github_actions_permissions" {
  statement {
    sid    = "ECRAuth"
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken",
    ]
    resources = ["*"] # this specific action doesn't support resource-level scoping
  }

  statement {
    sid    = "ECRPush"
    effect = "Allow"
    actions = [
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload",
      "ecr:PutImage",
    ]
    resources = [
      "arn:aws:ecr:${var.aws_region}:${data.aws_caller_identity.current.account_id}:repository/${var.project_name}-api",
      "arn:aws:ecr:${var.aws_region}:${data.aws_caller_identity.current.account_id}:repository/${var.project_name}-grpc-server",
    ]
  }

  statement {
    sid    = "EKSDescribe"
    effect = "Allow"
    actions = [
      "eks:DescribeCluster",
    ]
    resources = [module.eks.cluster_arn]
  }
}

resource "aws_iam_role_policy" "github_actions" {
  name   = "pipeline-permissions"
  role   = aws_iam_role.github_actions.id
  policy = data.aws_iam_policy_document.github_actions_permissions.json
}

# EKS access is separate from IAM -- having permission to call
# eks:DescribeCluster doesn't let you run kubectl commands. This grants the
# CI role edit rights (create/update/delete most objects, but not RBAC or
# node-level access) scoped to just this one namespace, not the whole
# cluster -- unlike the human admin access_entry above.
resource "aws_eks_access_entry" "github_actions" {
  cluster_name  = module.eks.cluster_name
  principal_arn = aws_iam_role.github_actions.arn
}

resource "aws_eks_access_policy_association" "github_actions" {
  cluster_name  = module.eks.cluster_name
  principal_arn = aws_iam_role.github_actions.arn
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSEditPolicy"

  access_scope {
    type       = "namespace"
    namespaces = ["satellite-tracker"]
  }

  # Both resources only reference aws_iam_role.github_actions.arn (a plain
  # string), so Terraform sees no resource-level edge between them and can
  # try to create this association before the access entry it depends on
  # actually exists -- exactly what happened on the first apply (404
  # ResourceNotFoundException). This forces the correct order.
  depends_on = [aws_eks_access_entry.github_actions]
}
