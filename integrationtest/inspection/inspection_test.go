package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
	"github.com/tofuutils/tofuenv/cmd"
	"github.com/tofuutils/tofuenv/formatter"
	"github.com/tofuutils/tofuenv/tofuutils"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

type meta struct {
	Version string
}

func TestIntegration(t *testing.T) {
	cases := []struct {
		Name    string
		Command string
		Env     map[string]string
		Dir     string
	}{
		{
			Name:    "basic",
			Command: "./tofulint --format json",
			Dir:     "basic",
		},
		{
			Name:    "override",
			Command: "./tofulint --format json",
			Dir:     "override",
		},
		{
			Name:    "variables",
			Command: "./tofulint --format json --var-file variables.tfvars --var var=var",
			Dir:     "variables",
		},
		{
			Name:    "module",
			Command: "./tofulint --format json --ignore-module ./ignore_module",
			Dir:     "module",
		},
		{
			Name:    "without module init",
			Command: "./tofulint --format json",
			Dir:     "without_module_init",
		},
		{
			Name:    "with module init",
			Command: "./tofulint --format json --call-module-type all",
			Dir:     "with_module_init",
		},
		{
			Name:    "no calling module",
			Command: "./tofulint --format json --call-module-type none",
			Dir:     "no_calling_module",
		},
		{
			Name:    "plugin",
			Command: "./tofulint --format json",
			Dir:     "plugin",
		},
		{
			Name:    "jsonsyntax",
			Command: "./tofulint --format json",
			Dir:     "jsonsyntax",
		},
		{
			Name:    "path",
			Command: "./tofulint --format json",
			Dir:     "path",
		},
		{
			Name:    "init from cwd",
			Command: "./tofulint --format json",
			Dir:     "init-cwd/root",
		},
		{
			Name:    "enable rule which has required configuration by CLI options",
			Command: "./tofulint --format json --enable-rule aws_s3_bucket_with_config_example",
			Dir:     "enable-required-config-rule-by-cli",
		},
		{
			Name:    "enable rule which does not have required configuration by CLI options",
			Command: "./tofulint --format json --enable-rule aws_db_instance_with_default_config_example",
			Dir:     "enable-config-rule-by-cli",
		},
		{
			Name:    "heredoc",
			Command: "./tofulint --format json",
			Dir:     "heredoc",
		},
		{
			Name:    "config parse error with HCL metadata",
			Command: "./tofulint --format json",
			Dir:     "bad-config",
		},
		{
			Name:    "conditional resources",
			Command: "./tofulint --format json",
			Dir:     "conditional",
		},
		{
			Name:    "dynamic blocks",
			Command: "./tofulint --format json",
			Dir:     "dynblock",
		},
		{
			Name:    "unknown dynamic blocks",
			Command: "./tofulint --format json",
			Dir:     "dynblock-unknown",
		},
		{
			Name:    "provider config",
			Command: "./tofulint --format json",
			Dir:     "provider-config",
		},
		{
			Name:    "rule config",
			Command: "./tofulint --format json",
			Dir:     "rule-config",
		},
		{
			Name:    "disabled rules",
			Command: "./tofulint --format json",
			Dir:     "disabled-rules",
		},
		{
			Name:    "cty-based eval",
			Command: "./tofulint --format json",
			Dir:     "cty-based-eval",
		},
		{
			Name:    "map attribute eval",
			Command: "./tofulint --format json",
			Dir:     "map-attribute",
		},
		{
			Name:    "rule config with --enable-rule",
			Command: "tofulint --enable-rule aws_s3_bucket_with_config_example --format json",
			Dir:     "rule-config",
		},
		{
			Name:    "rule config with --only",
			Command: "tofulint --only aws_s3_bucket_with_config_example --format json",
			Dir:     "rule-config",
		},
		{
			Name:    "rule config without required attributes",
			Command: "tofulint --format json",
			Dir:     "rule-required-config",
		},
		{
			Name:    "rule config without optional attributes",
			Command: "tofulint --format json",
			Dir:     "rule-optional-config",
		},
		{
			Name:    "enable plugin by CLI",
			Command: "tofulint --enable-plugin testing --format json",
			Dir:     "enable-plugin-by-cli",
		},
		{
			Name:    "eval on root context",
			Command: "tofulint --format json",
			Dir:     "eval-on-root-context",
		},
		{
			Name:    "sensitve variable",
			Command: "tofulint --format json",
			Dir:     "sensitive",
		},
		{
			Name:    "just attributes",
			Command: "tofulint --format json",
			Dir:     "just-attributes",
		},
		{
			Name:    "incompatible host version",
			Command: "tofulint --format json",
			Dir:     "incompatible-host",
		},
		{
			Name:    "expand resources/modules",
			Command: "tofulint --format json",
			Dir:     "expand",
		},
		{
			Name:    "chdir",
			Command: "tofulint --chdir dir --var-file from_cli.tfvars --format json",
			Dir:     "chdir",
		},
		{
			Name:    "recursive",
			Command: "tofulint --recursive --format json",
			Dir:     "recursive",
		},
	}

	// Disable the bundled plugin because the `os.Executable()` is go(1) in the tests
	tofulint.DisableBundledPlugin = true
	defer func() {
		tofulint.DisableBundledPlugin = false
	}()

	dir, _ := os.Getwd()
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			testDir := filepath.Join(dir, tc.Dir)

			defer func() {
				if err := os.Chdir(dir); err != nil {
					t.Fatal(err)
				}
			}()
			if err := os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}

			if tc.Env != nil {
				for k, v := range tc.Env {
					t.Setenv(k, v)
				}
			}

			outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
			cli, err := cmd.NewCLI(outStream, errStream)
			if err != nil {
				t.Fatal(err)
			}
			args := strings.Split(tc.Command, " ")

			cli.Run(args)

			rawWant, err := readResultFile(testDir)
			if err != nil {
				t.Fatal(err)
			}
			var want *formatter.JSONOutput
			if err := json.Unmarshal(rawWant, &want); err != nil {
				t.Fatal(err)
			}

			var got *formatter.JSONOutput
			if err := json.Unmarshal(outStream.Bytes(), &got); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(got, want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func readResultFile(dir string) ([]byte, error) {
	resultFile := "result.json"
	if runtime.GOOS == "windows" {
		if _, err := os.Stat(filepath.Join(dir, "result_windows.json")); !os.IsNotExist(err) {
			resultFile = "result_windows.json"
		}
	}
	if _, err := os.Stat(fmt.Sprintf("%s.tmpl", resultFile)); !os.IsNotExist(err) {
		resultFile = fmt.Sprintf("%s.tmpl", resultFile)
	}

	if !strings.HasSuffix(resultFile, ".tmpl") {
		return os.ReadFile(filepath.Join(dir, resultFile))
	}

	want := new(bytes.Buffer)
	tmpl := template.Must(template.ParseFiles(filepath.Join(dir, resultFile)))
	if err := tmpl.Execute(want, meta{Version: tofulint.Version.String()}); err != nil {
		return nil, err
	}
	return want.Bytes(), nil
}
