package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl"
)

// HclVar is a parsed HCL variable.
type HclVar struct {
	Name        string
	Description string
	VarType     string
	DefaultVal  string
	Required    bool
	Sensitive   bool
}

const (
	variablesFilename = "variables.tf"
	outputFilename    = "outputs.tf"
)

var (
	verboseFlag       bool   // Verbose mode.
	variablesFileFlag string // Path to variables file.
	outputsFileFlag   string // Path to outputs file.
	lg                cliLogger
)

func init() {
	flag.BoolVar(&verboseFlag, "v", false, "verbose mode")
	flag.StringVar(&variablesFileFlag, "variables", "", "path to variables file")
	flag.StringVar(&outputsFileFlag, "outputs", "", "path to outputs file")
	flag.Parse()
}

// exists returns false if the given file does not exist, true otherwise.
func exists(name string) ([]byte, bool) {
	if _, err := os.Stat(variablesFilename); err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
	}
	out, err := ioutil.ReadFile("./" + name)
	if err != nil {
		lg.debugf("Error reading %s: %s", name, err)
		return nil, false
	}
	return out, true
}

func printTitle() {
	// Construct the title header.
	wd, err := os.Getwd()
	if err != nil {
		lg.Fatalf("Error getting working dir: %s", err)
	}
	fmt.Printf("\n# %s Terraform Module\n", strings.ToTitle(filepath.Base(wd)))
}

func main() {
	printTitle()

	// Overview.
	fmt.Printf("\n## Overview\n\n")

	// Handle input variables.
	if rawInputs, ok := exists(variablesFilename); ok {
		var hclInput interface{}
		if err := hcl.Unmarshal(rawInputs, &hclInput); err != nil {
			lg.Fatalf("Error unmarshalling input: %s", err)
		}

		vars, ok := hclInput.(map[string]interface{})["variable"]
		if !ok && verboseFlag {
			lg.Printf("No variables detected.")
		}

		hclVars := make([]HclVar, len(vars.([]map[string]interface{})))
		var desc, varType, defaultVal string
		for varindex, varmap := range vars.([]map[string]interface{}) {
			for name, v := range varmap {
				for _, x := range v.([]map[string]interface{}) {
					desc, _ = x["description"].(string)
					varType, _ = x["type"].(string)
					defaultVal, _ = x["default"].(string)
					hclvar := HclVar{
						Name:        name,
						Description: desc,
						VarType:     varType,
						DefaultVal:  defaultVal,
					}
					if defaultVal != "" {
						hclvar.Required = true
					} else {
						hclvar.Required = false
					}
					hclVars[varindex] = hclvar
				}
			}
		}

		// Format and print Inputs.
		inputTmpl, err := template.New("hclvar_input").Parse("| {{.Name}} | {{.Description}} | {{.VarType}} | {{.DefaultVal}} | {{if .Required}} yes {{else}} no {{end}} |\n")
		if err != nil {
			lg.Fatalf("Error templating input: %s", err)
		}
		fmt.Printf("\n## Input\n\n")
		fmt.Println("| Name | Description | Type | Default | Required |")
		fmt.Println("|------|-------------|:----:|:-----:|:-----:|")
		for _, hclvar := range hclVars {
			err = inputTmpl.Execute(os.Stdout, hclvar)
			if err != nil {
				lg.Fatalf("Error executing input on template: %s", err)
			}
		}
	}

	// Handle outputs.
	if rawOutputs, ok := exists(outputFilename); ok {
		var hclOut interface{}
		if err := hcl.Unmarshal(rawOutputs, &hclOut); err != nil {
			lg.Fatalf("Error unmarshalling: %s", err)
		}

		outputs, ok := hclOut.(map[string]interface{})["output"]
		if !ok && verboseFlag {
			lg.Printf("No outputs detected.")
		}

		hclOutputs := make([]HclVar, len(outputs.([]map[string]interface{})))
		var outputDesc string
		var outputIsSensitive bool
		for outindex, outmap := range outputs.([]map[string]interface{}) {
			for name, v := range outmap {
				for _, x := range v.([]map[string]interface{}) {
					outputDesc, _ = x["description"].(string)
					outputIsSensitive, _ = x["sensitive"].(bool)
					hclvar := HclVar{
						Name:        name,
						Description: outputDesc,
						Sensitive:   outputIsSensitive,
					}
					hclOutputs[outindex] = hclvar
				}
			}
		}

		// Format and print Outputs.
		outputTmpl, err := template.New("hclvar_output").Parse("| {{.Name}} | {{.Description}} |  {{if .Sensitive}} yes {{else}} no {{end}} |\n")
		if err != nil {
			lg.Fatalf("Error templating output: %s", err)
		}
		fmt.Printf("\n## Output\n\n")
		fmt.Println("| Name | Description | Sensitive |")
		fmt.Println("|------|-------------|:----:|")
		for _, out := range hclOutputs {
			err = outputTmpl.Execute(os.Stdout, out)
			if err != nil {
				lg.Fatalf("Error executing output on template: %s", err)
			}
		}
	}

	// Usage.
	fmt.Printf("\n## Usage\n")
	fmt.Printf("\n```\n\n```\n")

	// Troubleshooting.
	fmt.Printf("\n## Troubleshooting\n\n")
}
