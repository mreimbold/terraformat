module "vpc" {
  version = "5.0.0"
  source  = "terraform-aws-modules/vpc/aws"
}

resource "aws_instance" "app" {
  depends_on = [aws_security_group.app]
  ami        = "ami-123"

  lifecycle {
    prevent_destroy = true
  }

  provisioner "local-exec" {
    command = "echo hi"
  }

  tags = {
    Name = "app"
  }

  count    = 2
  provider = aws.primary

  network_interface {
    device_index = 0
    network_interface_id = aws_network_interface.app.id
  }

  provisioner "local-exec" {
    command = "echo bye"
  }
}

terraform {
  backend "s3" {
    bucket = "state"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  required_version = ">= 1.6.0"
}

# comment for variables
variable "region" {
  default = "us-east-1"
  type    = string
  description = "AWS region"

  validation {
    condition     = length(var.region) > 0
    error_message = "Region is required."
  }
}

variable "enabled" {
  description = "Toggle"
  default     = true
  type        = bool
}

provider "aws" {
  region = var.region
  alias  = "primary"
}

locals {
  tags = {
    Env = "dev"
  }
  name = "app"
}

data "aws_caller_identity" "current" {}

output "instance_id" {
  value       = aws_instance.app.id
  sensitive   = false
  description = "Instance ID"
}

resource "aws_security_group" "app" {
  name = "app"

  ingress {
    from_port = 443
    to_port   = 443
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}
