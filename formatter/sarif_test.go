package formatter

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/tofuutils/tofulint/tofulint"
	"github.com/xeipuuv/gojsonschema"
)

func Test_sarifPrint(t *testing.T) {
	cases := []struct {
		Name   string
		Issues tofulint.Issues
		Error  error
		Stdout string
	}{
		{
			Name:   "no issues",
			Issues: tofulint.Issues{},
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tofulint",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint"
        }
      },
      "results": []
    },
    {
      "tool": {
        "driver": {
          "name": "tofulint-errors",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint"
        }
      },
      "results": []
    }
  ]
}`, tofulint.Version, tofulint.Version),
		},
		{
			Name: "issues",
			Issues: tofulint.Issues{
				{
					Rule:    &testRule{},
					Message: "test",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tofulint",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint",
          "rules": [
            {
              "id": "test_rule",
              "shortDescription": {
                "text": ""
              },
              "helpUri": "https://github.com"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "test_rule",
          "level": "error",
          "message": {
            "text": "test"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.tf"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 4
                }
              }
            }
          ]
        }
      ]
    },
    {
      "tool": {
        "driver": {
          "name": "tofulint-errors",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint"
        }
      },
      "results": []
    }
  ]
}`, tofulint.Version, tofulint.Version),
		},
		{
			Name: "issues not on line 1",
			Issues: tofulint.Issues{
				{
					Rule:    &testRule{},
					Message: "test",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 3, Column: 4, Byte: 3},
					},
				},
			},
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tofulint",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint",
          "rules": [
            {
              "id": "test_rule",
              "shortDescription": {
                "text": ""
              },
              "helpUri": "https://github.com"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "test_rule",
          "level": "error",
          "message": {
            "text": "test"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.tf"
                },
                "region": {
                  "startLine": 3,
                  "startColumn": 1,
                  "endLine": 3,
                  "endColumn": 4
                }
              }
            }
          ]
        }
      ]
    },
    {
      "tool": {
        "driver": {
          "name": "tofulint-errors",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint"
        }
      },
      "results": []
    }
  ]
}`, tofulint.Version, tofulint.Version),
		},
		{
			Name: "issues spanning multiple lines",
			Issues: tofulint.Issues{
				{
					Rule:    &testRule{},
					Message: "test",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 4, Column: 1, Byte: 3},
					},
				},
			},
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tofulint",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint",
          "rules": [
            {
              "id": "test_rule",
              "shortDescription": {
                "text": ""
              },
              "helpUri": "https://github.com"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "test_rule",
          "level": "error",
          "message": {
            "text": "test"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.tf"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 4,
                  "endColumn": 1
                }
              }
            }
          ]
        }
      ]
    },
    {
      "tool": {
        "driver": {
          "name": "tofulint-errors",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint"
        }
      },
      "results": []
    }
  ]
}`, tofulint.Version, tofulint.Version),
		},
		{
			Name: "issues in directories",
			Issues: tofulint.Issues{
				{
					Rule:    &testRule{},
					Message: "test",
					Range: hcl.Range{
						Filename: filepath.Join("test", "main.tf"),
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tofulint",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint",
          "rules": [
            {
              "id": "test_rule",
              "shortDescription": {
                "text": ""
              },
              "helpUri": "https://github.com"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "test_rule",
          "level": "error",
          "message": {
            "text": "test"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test/main.tf"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 4
                }
              }
            }
          ]
        }
      ]
    },
    {
      "tool": {
        "driver": {
          "name": "tofulint-errors",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint"
        }
      },
      "results": []
    }
  ]
}`, tofulint.Version, tofulint.Version),
		},
		{
			Name: "Issues with missing source positions",
			Issues: tofulint.Issues{
				{
					Rule:    &testRule{},
					Message: "test",
					Range: hcl.Range{
						Filename: "test.tf",
					},
				},
			},
			Error: fmt.Errorf("Failed to work; %w", errors.New("I don't feel like working")),
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tofulint",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint",
          "rules": [
            {
              "id": "test_rule",
              "shortDescription": {
                "text": ""
              },
              "helpUri": "https://github.com"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "test_rule",
          "level": "error",
          "message": {
            "text": "test"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.tf"
                }
              }
            }
          ]
        }
      ]
    },
    {
      "tool": {
        "driver": {
          "name": "tofulint-errors",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint"
        }
      },
      "results": [
        {
          "ruleId": "application_error",
          "level": "error",
          "message": {
            "text": "Failed to work; I don't feel like working"
          }
        }
      ]
    }
  ]
}`, tofulint.Version, tofulint.Version),
		},
		{
			Name: "HCL diagnostics are surfaced as tofulint-errors",
			Error: fmt.Errorf(
				"babel fish confused; %w",
				hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagWarning,
						Summary:  "summary",
						Detail:   "detail",
						Subject: &hcl.Range{
							Filename: "filename",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 5, Column: 1, Byte: 4},
						},
					},
				},
			),
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tofulint",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint"
        }
      },
      "results": []
    },
    {
      "tool": {
        "driver": {
          "name": "tofulint-errors",
          "version": "%s",
          "informationUri": "https://github.com/tofuutils/tofulint"
        }
      },
      "results": [
        {
          "ruleId": "summary",
          "level": "warning",
          "message": {
            "text": "detail"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "filename"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 5,
                  "endColumn": 1,
                  "byteOffset": 0,
                  "byteLength": 4
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`, tofulint.Version, tofulint.Version),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			formatter := &Formatter{Stdout: stdout, Stderr: stderr, Format: "sarif"}

			formatter.Print(tc.Issues, tc.Error, map[string][]byte{})

			if diff := cmp.Diff(tc.Stdout, stdout.String()); diff != "" {
				t.Fatalf("Failed %s test: %s", tc.Name, diff)
			}

			schemaLoader := gojsonschema.NewReferenceLoader("http://json.schemastore.org/sarif-2.1.0")
			result, err := gojsonschema.Validate(schemaLoader, gojsonschema.NewStringLoader(stdout.String()))
			if err != nil {
				t.Error(err)
			}
			for _, err := range result.Errors() {
				t.Error(err)
			}
		})
	}
}
