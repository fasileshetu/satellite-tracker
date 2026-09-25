# AWS CodeArtifact is the JD's "binary management" line -- a private
# artifact repository, separate from ECR (which only stores container
# images). CodeArtifact natively understands npm/PyPI/Maven/NuGet package
# formats, but also has a "generic" format for arbitrary files, which is
# what a compiled Go binary is. The pipeline publishes the built `api` and
# `grpc-server` binaries here on every push to main, giving a versioned,
# access-controlled history of exactly what was built -- independent of
# (and before) it ever becomes a container image.
resource "aws_codeartifact_domain" "this" {
  domain = var.project_name
}

resource "aws_codeartifact_repository" "binaries" {
  repository = "binaries"
  domain     = aws_codeartifact_domain.this.domain
  description = "Compiled Go binaries (api, grpc-server) published by CI on every push to main"
}
