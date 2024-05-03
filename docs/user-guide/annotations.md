# Annotations

Annotation comments can disable rules on specific lines:

```hcl
resource "aws_instance" "foo" {
    # tofulint-ignore: aws_instance_invalid_type
    instance_type = "t1.2xlarge"
}
```

Multiple rules can be specified as a comma-separated list:

```hcl
resource "aws_instance" "foo" {
    # tofulint-ignore: aws_instance_invalid_type, other_rule
    instance_type = "t1.2xlarge"
}
```

All rules can be ignored by specifying the `all` keyword:

```hcl
resource "aws_instance" "foo" {
    # tofulint-ignore: all
    instance_type = "t1.2xlarge"
}
```

It's a good idea to add a reason for why a rule is ignored, especially temporarily:

```hcl
resource "aws_instance" "foo" {
    # This instance type is new and TofuLint doesn't know about it yet
    # tofulint-ignore: aws_instance_invalid_type
    instance_type = "t10.2xlarge"
}
```

Or, on the same line:

```hcl
resource "aws_instance" "foo" {
  # tofulint-ignore: aws_instance_invalid_type # too new for TofuLint
  instance_type = "t10.2xlarge" 
}
```

The `//` comment style is also supported, but Terraform recommends `#`.

```hcl
resource "aws_instance" "foo" {
  // tofulint-ignore: aws_instance_invalid_type // too new for TofuLint
  instance_type = "t10.2xlarge" 
}
```

To disable an entire file, you can also use the `tofulint-ignore-file` annotation:

```hcl
# tofulint-ignore-file: aws_instance_invalid_type

resource "aws_instance" "foo" {
  instance_type = "t1.2xlarge"
}
```

This annotation is valid only at the top of the file. The following cannot be used and will result in an error:

```hcl
resource "aws_instance" "foo" {
  # tofulint-ignore-file: aws_instance_invalid_type
  instance_type = "t1.2xlarge"
}
```

```hcl
resource "aws_instance" "foo" { # tofulint-ignore-file: aws_instance_invalid_type
  instance_type = "t1.2xlarge"
}
```
