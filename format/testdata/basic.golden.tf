# comment for variable
variable "name" {
  type        = string
  description = "Name"
  default     = "web"
}

resource "aws_instance" "a" {
  provider = aws.foo

  ami = "ami-a"
  tags = {
    Name = "a"
  }

  depends_on = [aws_instance.b]
}

resource "aws_instance" "b" {
  count = 1

  ami = "ami-b"

  lifecycle {
    prevent_destroy = true
  }

  depends_on = [aws_vpc.main]
}
