terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket = "state"
  }
}

provider "aws" {
  alias  = "primary"
  region = var.region
}

variable "enabled" {
  type        = bool
  description = "Toggle"
  default     = true
}

# comment for variables
variable "region" {
  type        = string
  description = "AWS region"
  default     = "us-east-1"

  validation {
    condition     = length(var.region) > 0
    error_message = "Region is required."
  }
}

locals {
  tags = {
    Env = "dev"
  }
  name = "app"
}

data "aws_caller_identity" "current" {}

resource "aws_instance" "app" {
  count    = 2
  provider = aws.primary

  ami = "ami-123"
  tags = {
    Name = "app"
  }

  provisioner "local-exec" {
    command = "echo hi"
  }

  network_interface {
    device_index         = 0
    network_interface_id = aws_network_interface.app.id
  }

  provisioner "local-exec" {
    command = "echo bye"
  }

  lifecycle {
    prevent_destroy = true
  }

  depends_on = [aws_security_group.app]
}

resource "aws_security_group" "app" {
  name = "app"

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"
}

output "instance_id" {
  description = "Instance ID"
  value       = aws_instance.app.id
  sensitive   = false
}
