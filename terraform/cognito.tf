# Cognito is the OIDC identity provider for this project. Unlike EKS/RDS,
# it has no meaningful ongoing cost (the free tier covers far more monthly
# active users than a portfolio project will ever see) and doesn't depend
# on the rest of the stack at all -- it can be applied and tested on its
# own even with the cluster torn down.

resource "aws_cognito_user_pool" "this" {
  name = "${var.project_name}-users"

  # Cognito's own hosted sign-up/sign-in UI needs the username to be an
  # email, and auto-verifying it keeps this simple for a demo/portfolio
  # setup (a real production pool would require verification).
  username_attributes     = ["email"]
  auto_verified_attributes = ["email"]

  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_numbers   = true
    require_symbols   = false
    require_uppercase = true
  }

  tags = {
    Project = var.project_name
  }
}

# The "app client" is the OAuth client_id your API/dashboard authenticates
# as. generate_secret = false because this is a "public" client (a
# frontend or CLI, nothing that can safely keep a secret) -- the correct
# choice for the Authorization Code + PKCE flow a real frontend would use.
resource "aws_cognito_user_pool_client" "this" {
  name         = "${var.project_name}-client"
  user_pool_id = aws_cognito_user_pool.this.id

  generate_secret = false

  # USER_PASSWORD_AUTH lets you get a real token straight from the AWS
  # CLI for testing the Go middleware, without standing up a frontend
  # first. ALLOW_REFRESH_TOKEN_AUTH is the standard companion to it.
  explicit_auth_flows = [
    "ALLOW_USER_PASSWORD_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH",
  ]

  # These matter once a real frontend does the Authorization Code + PKCE
  # flow (the production pattern) instead of the CLI shortcut above.
  allowed_oauth_flows                 = ["code"]
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_scopes                = ["openid", "email", "profile"]
  supported_identity_providers        = ["COGNITO"]
  callback_urls                       = ["http://localhost:3000/callback"]
  logout_urls                         = ["http://localhost:3000"]

  access_token_validity  = 60 # minutes
  id_token_validity      = 60
  refresh_token_validity = 30 # days

  token_validity_units {
    access_token  = "minutes"
    id_token      = "minutes"
    refresh_token = "days"
  }
}

# Cognito's Hosted UI -- a login page AWS provides for free, so a real
# Authorization Code redirect flow can be demoed without building a
# custom login page first. Needed for allowed_oauth_flows above to work.
#
# The domain prefix must be globally unique across every AWS account, not
# just yours -- it becomes <domain>.auth.<region>.amazoncognito.com -- so
# a random suffix avoids colliding with someone else's "satellite-tracker".
resource "random_id" "cognito_domain_suffix" {
  byte_length = 4
}

resource "aws_cognito_user_pool_domain" "this" {
  domain       = "${var.project_name}-auth-${random_id.cognito_domain_suffix.hex}"
  user_pool_id = aws_cognito_user_pool.this.id
}
