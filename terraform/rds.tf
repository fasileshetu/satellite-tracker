# Written as plain resources rather than a module, since it's small enough
# to read start to finish and worth seeing exactly what RDS needs.

resource "aws_db_subnet_group" "this" {
  name       = "${var.project_name}-db-subnets"
  subnet_ids = module.vpc.private_subnets

  tags = {
    Project = var.project_name
  }
}

# Only allow Postgres traffic (port 5432) in from the EKS worker nodes'
# security group — nothing else on the internet can reach this database.
resource "aws_security_group" "rds" {
  name        = "${var.project_name}-rds-sg"
  description = "Allow Postgres access from EKS nodes only"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description     = "Postgres from EKS nodes"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [module.eks.node_security_group_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Project = var.project_name
  }
}

resource "aws_db_instance" "this" {
  identifier     = "${var.project_name}-db"
  engine         = "postgres"
  engine_version = "16"

  instance_class    = var.db_instance_class
  allocated_storage = 20 # GB — the minimum, free-tier eligible

  db_name  = var.db_name
  username = var.db_username
  password = var.db_password

  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  publicly_accessible = false # only reachable from inside the VPC, i.e. from the EKS nodes
  skip_final_snapshot = true  # fine for a dev/portfolio project; a real prod DB would not set this
  multi_az             = false # single instance — keeps cost down, no failover redundancy

  tags = {
    Project = var.project_name
  }
}
