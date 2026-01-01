# comment for variable
variable "name" {
  type        = string
  description = "Name"
  default     = "web"
}

resource "aws_instance" "b" {
  count = 1

  ami = "ami-b"

  lifecycle {
    prevent_destroy = true
  }

  depends_on = [aws_vpc.main]
}

resource "aws_instance" "a" {
  provider = aws.foo

  tags = {
    Name = "a"
  }
  ami = "ami-a"

  depends_on = [aws_instance.b]
}
