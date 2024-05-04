resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}

resource "aws_instance" "bar" {
  // tofulint-ignore: aws_instance_example_type
  instance_type = "t2.micro"
}
