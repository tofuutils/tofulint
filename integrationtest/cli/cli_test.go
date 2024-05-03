package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/tofuutils/tofulint/cmd"
	"github.com/tofuutils/tofulint/tofulint"
)

func TestIntegration(t *testing.T) {
	// Disable the bundled plugin because the `os.Executable()` is go(1) in the tests
	tofulint.DisableBundledPlugin = true
	defer func() {
		tofulint.DisableBundledPlugin = false
	}()

	tests := []struct {
		name    string
		command string
		dir     string
		status  int
		stdout  string
		stderr  string
	}{
		{
			name:    "print version",
			command: "./tofulint --version",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  fmt.Sprintf("TofuLint version %s", tofulint.Version),
		},
		{
			name:    "print help",
			command: "./tofulint --help",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "Application Options:",
		},
		{
			name:    "no options",
			command: "./tofulint",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "",
		},
		{
			name:    "specify format",
			command: "./tofulint --format json",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "[]",
		},
		{
			name:    "`--force` option with no issues",
			command: "./tofulint --force",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "",
		},
		{
			name:    "`--minimum-failure-severity` option with no issues",
			command: "./tofulint --minimum-failure-severity=notice",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "",
		},
		{
			name:    "`--only` option",
			command: "./tofulint --only aws_instance_example_type",
			dir:     "no_issues",
			status:  cmd.ExitCodeOK,
			stdout:  "",
		},
		{
			name:    "loading errors are occurred",
			command: "./tofulint",
			dir:     "load_errors",
			status:  cmd.ExitCodeError,
			stderr:  "Failed to load configurations;",
		},
		{
			name:    "removed --debug options",
			command: "./tofulint --debug",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--debug option was removed in v0.8.0. Please set TOFULINT_LOG environment variables instead",
		},
		{
			name:    "removed --error-with-issues option",
			command: "./tofulint --error-with-issues",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--error-with-issues option was removed in v0.9.0. The behavior is now default",
		},
		{
			name:    "removed --quiet option",
			command: "./tofulint --quiet",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--quiet option was removed in v0.11.0. The behavior is now default",
		},
		{
			name:    "removed --ignore-rule option",
			command: "./tofulint --ignore-rule aws_instance_example_type",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--ignore-rule option was removed in v0.12.0. Please use --disable-rule instead",
		},
		{
			name:    "removed --deep option",
			command: "./tofulint --deep",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--deep option was removed in v0.23.0. Deep checking is now a feature of the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --aws-access-key option",
			command: "./tofulint --aws-access-key AWS_ACCESS_KEY_ID",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--aws-access-key option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --aws-secret-key option",
			command: "./tofulint --aws-secret-key AWS_SECRET_ACCESS_KEY",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--aws-secret-key option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --aws-profile option",
			command: "./tofulint --aws-profile AWS_PROFILE",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--aws-profile option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --aws-creds-file option",
			command: "./tofulint --aws-creds-file FILE",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--aws-creds-file option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --aws-region option",
			command: "./tofulint --aws-region us-east-1",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--aws-region option was removed in v0.23.0. AWS rules are provided by the AWS plugin, so please configure the plugin instead",
		},
		{
			name:    "removed --loglevel option",
			command: "./tofulint --loglevel debug",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "--loglevel option was removed in v0.40.0. Please set TOFULINT_LOG environment variables instead",
		},
		{
			name:    "invalid options",
			command: "./tofulint --unknown",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  `--unknown is unknown option. Please run "tofulint --help"`,
		},
		{
			name:    "invalid format",
			command: "./tofulint --format awesome",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "Invalid value `awesome' for option",
		},
		{
			name:    "invalid rule name",
			command: "./tofulint --enable-rule nosuchrule",
			dir:     "no_issues",
			status:  cmd.ExitCodeError,
			stderr:  "Rule not found: nosuchrule",
		},
		{
			name:    "issues found",
			command: "./tofulint",
			dir:     "issues_found",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is t2.micro")),
		},
		{
			name:    "--force option with issues",
			command: "./tofulint --force",
			dir:     "issues_found",
			status:  cmd.ExitCodeOK,
			stdout:  fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is t2.micro")),
		},
		{
			name:    "--minimum-failure-severity option with warning issues and minimum-failure-severity notice",
			command: "./tofulint --minimum-failure-severity=notice",
			dir:     "warnings_found",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_s3_bucket_with_config_example)", color.New(color.Bold).Sprint("bucket name is test, config=bucket")),
		},
		{
			name:    "--minimum-failure-severity option with warning issues and minimum-failure-severity warning",
			command: "./tofulint --minimum-failure-severity=warning",
			dir:     "warnings_found",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_s3_bucket_with_config_example)", color.New(color.Bold).Sprint("bucket name is test, config=bucket")),
		},
		{
			name:    "--minimum-failure-severity option with warning issues and minimum-failure-severity error",
			command: "./tofulint --minimum-failure-severity=error",
			dir:     "warnings_found",
			status:  cmd.ExitCodeOK,
			stdout:  fmt.Sprintf("%s (aws_s3_bucket_with_config_example)", color.New(color.Bold).Sprint("bucket name is test, config=bucket")),
		},
		{
			name:    "--no-color option",
			command: "./tofulint --no-color",
			dir:     "issues_found",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  "instance type is t2.micro (aws_instance_example_type)",
		},
		{
			name:    "checking errors are occurred",
			command: "./tofulint",
			dir:     "check_errors",
			status:  cmd.ExitCodeError,
			stderr:  `failed to check "aws_cloudformation_stack_error" rule: an error occurred in Check`,
		},
		{
			name:    "files arguments",
			command: "./tofulint empty.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "file not found",
			command: "./tofulint not_found.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "not OpenTofu configuration",
			command: "./tofulint README.md",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "multiple files",
			command: "./tofulint empty.tf main.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "directory argument",
			command: "./tofulint subdir",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "file under the directory",
			command: fmt.Sprintf("./tofulint %s", filepath.Join("subdir", "main.tf")),
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "multiple directories",
			command: "./tofulint subdir ./",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "file and directory",
			command: "./tofulint main.tf subdir",
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "multiple files in different directories",
			command: fmt.Sprintf("./tofulint main.tf %s", filepath.Join("subdir", "main.tf")),
			dir:     "multiple_files",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--filter",
			command: "./tofulint --filter=empty.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeOK,
			stdout:  "", // main.tf is ignored
		},
		{
			name:    "--filter with multiple files",
			command: "./tofulint --filter=empty.tf --filter=main.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeIssuesFound,
			// main.tf is not ignored
			stdout: fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is t2.micro")),
		},
		{
			name:    "--filter with glob (files found)",
			command: "./tofulint --filter=*.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is t2.micro")),
		},
		{
			name:    "--filter with glob (files not found)",
			command: "./tofulint --filter=*_generated.tf",
			dir:     "multiple_files",
			status:  cmd.ExitCodeOK,
			stdout:  "",
		},
		{
			name:    "--chdir",
			command: "./tofulint --chdir=subdir",
			dir:     "chdir",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is m5.2xlarge")),
		},
		{
			name:    "--chdir and file argument",
			command: "./tofulint --chdir=subdir main.tf",
			dir:     "chdir",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--chdir and directory argument",
			command: "./tofulint --chdir=subdir ../",
			dir:     "chdir",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--chdir and the current directory argument",
			command: "./tofulint --chdir=subdir .",
			dir:     "chdir",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--chdir and file under the directory argument",
			command: fmt.Sprintf("./tofulint --chdir=subdir %s", filepath.Join("nested", "main.tf")),
			dir:     "chdir",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--chdir and --filter",
			command: "./tofulint --chdir=subdir --filter=main.tf",
			dir:     "chdir",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is m5.2xlarge")),
		},
		{
			name:    "--recursive and file argument",
			command: "./tofulint --recursive main.tf",
			dir:     "chdir",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--recursive and directory argument",
			command: "./tofulint --recursive subdir",
			dir:     "chdir",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--recursive and the current directory argument",
			command: "./tofulint --recursive .",
			dir:     "chdir",
			status:  cmd.ExitCodeError,
			stderr:  `Command line arguments support was dropped in v0.47. Use --chdir or --filter instead.`,
		},
		{
			name:    "--recursive and --filter",
			command: "./tofulint --recursive --filter=main.tf",
			dir:     "chdir",
			status:  cmd.ExitCodeIssuesFound,
			stdout:  fmt.Sprintf("%s (aws_instance_example_type)", color.New(color.Bold).Sprint("instance type is m5.2xlarge")),
		},
	}

	dir, _ := os.Getwd()
	defaultNoColor := color.NoColor

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testDir := filepath.Join(dir, test.dir)
			defer func() {
				if err := os.Chdir(dir); err != nil {
					t.Fatal(err)
				}
				// Reset global color option
				color.NoColor = defaultNoColor
			}()
			if err := os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}

			outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
			cli, err := cmd.NewCLI(outStream, errStream)
			if err != nil {
				t.Fatal(err)
			}
			args := strings.Split(test.command, " ")

			got := cli.Run(args)

			if got != test.status {
				t.Errorf("expected status is %d, but got %d", test.status, got)
			}
			if !strings.Contains(outStream.String(), test.stdout) || (test.stdout == "" && outStream.String() != "") {
				t.Errorf("stdout did not contain expected\n\texpected: %s\n\tgot: %s", test.stdout, outStream.String())
			}
			if !strings.Contains(errStream.String(), test.stderr) || (test.stderr == "" && errStream.String() != "") {
				t.Errorf("stderr did not contain expected\n\texpected: %s\n\tgot: %s", test.stderr, errStream.String())
			}
		})
	}
}
