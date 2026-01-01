resource "aws_instance" "b" {
  ami = "ami-b"
  depends_on = [aws_vpc.main]

  lifecycle {
    prevent_destroy = true
  }

  count = 1
}

# comment for variable


variable "name" {
  default     = "web"
  description = "Name"
  type        = string
}

resource "aws_instance" "a" {
  tags = {
    Name = "a"
  }
  provider   = aws.foo
  ami        = "ami-a"
  depends_on = [aws_instance.b]
}
