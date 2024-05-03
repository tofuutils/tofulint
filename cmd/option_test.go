package cmd

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	flags "github.com/jessevdk/go-flags"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/tofuutils/tofulint/tofulint"
)

func Test_toConfig(t *testing.T) {
	cases := []struct {
		Name     string
		Command  string
		Expected *tofulint.Config
	}{
		{
			Name:     "default",
			Command:  "./tofulint",
			Expected: tflint.EmptyConfig(),
		},
		{
			Name:    "--call-module-type",
			Command: "./tofulint --call-module-type all",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallAllModule,
				CallModuleTypeSet: true,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins:           map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "--module",
			Command: "./tofulint --module",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallAllModule,
				CallModuleTypeSet: true,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins:           map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "--no-module",
			Command: "./tofulint --no-module",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallNoModule,
				CallModuleTypeSet: true,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins:           map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "--module and --call-module-type",
			Command: "./tofulint --module --call-module-type none",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallNoModule,
				CallModuleTypeSet: true,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins:           map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "--force",
			Command: "./tofulint --force",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             true,
				ForceSet:          true,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins:           map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "--ignore-module",
			Command: "./tofulint --ignore-module module1,module2",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{"module1": true, "module2": true},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins:           map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "multiple --ignore-module",
			Command: "./tofulint --ignore-module module1 --ignore-module module2",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{"module1": true, "module2": true},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins:           map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "--var-file",
			Command: "./tofulint --var-file example1.tfvars,example2.tfvars",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{"example1.tfvars", "example2.tfvars"},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins:           map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "multiple --var-file",
			Command: "./tofulint --var-file example1.tfvars --var-file example2.tfvars",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{"example1.tfvars", "example2.tfvars"},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins:           map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "--var",
			Command: "./tofulint --var foo=bar --var bar=baz",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{"foo=bar", "bar=baz"},
				DisabledByDefault: false,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins:           map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "--enable-rule",
			Command: "./tofulint --enable-rule aws_instance_invalid_type --enable-rule aws_instance_previous_type",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules: map[string]*tofulint.RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: true,
						Body:    nil,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: true,
						Body:    nil,
					},
				},
				Plugins: map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "--disable-rule",
			Command: "./tofulint --disable-rule aws_instance_invalid_type --disable-rule aws_instance_previous_type",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules: map[string]*tofulint.RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: false,
						Body:    nil,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: false,
						Body:    nil,
					},
				},
				Plugins: map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "--only",
			Command: "./tofulint --only aws_instance_invalid_type",
			Expected: &tflint.Config{
				CallModuleType:       terraform.CallLocalModule,
				Force:                false,
				IgnoreModules:        map[string]bool{},
				Varfiles:             []string{},
				Variables:            []string{},
				DisabledByDefault:    true,
				DisabledByDefaultSet: true,
				Only:                 []string{"aws_instance_invalid_type"},
				Rules: map[string]*tofulint.RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: true,
						Body:    nil,
					},
				},
				Plugins: map[string]*tofulint.PluginConfig{},
			},
		},
		{
			Name:    "--enable-plugin",
			Command: "./tofulint --enable-plugin test --enable-plugin another-test",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins: map[string]*tofulint.PluginConfig{
					"test": {
						Name:    "test",
						Enabled: true,
						Body:    nil,
					},
					"another-test": {
						Name:    "another-test",
						Enabled: true,
						Body:    nil,
					},
				},
			},
		},
		{
			Name:    "--format",
			Command: "./tofulint --format compact",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Format:            "compact",
				FormatSet:         true,
				Rules:             map[string]*tofulint.RuleConfig{},
				Plugins:           map[string]*tofulint.PluginConfig{},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var opts Options
			parser := flags.NewParser(&opts, flags.HelpFlag)

			_, err := parser.ParseArgs(strings.Split(tc.Command, " "))
			if err != nil {
				t.Fatal(err)
			}

			ret := opts.toConfig()
			eqlopts := []cmp.Option{
				cmpopts.IgnoreUnexported(tflint.RuleConfig{}),
				cmpopts.IgnoreUnexported(tflint.PluginConfig{}),
				cmpopts.IgnoreUnexported(tflint.Config{}),
			}
			if diff := cmp.Diff(tc.Expected, ret, eqlopts...); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
