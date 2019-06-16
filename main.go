package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl"
)

// HCLVar is a parsed HCL variable.
type HCLVar struct {
	Name        string
	Description string
	VarType     string
	DefaultVal  string
	Required    bool
	Sensitive   bool
}

func printTitle() {
	// Construct the title header.
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working dir: %s", err)
	}
	fmt.Printf("\n# %s Terraform Module\n", strings.ToTitle(filepath.Base(wd)))
}

func main() {
	var (
		verbose       = flag.Bool("v", false, "verbose mode")
		variablesFile = flag.String("variables", "variables.tf", "path to variables file")
		outputsFile   = flag.String("outputs", "outputs.tf", "path to outputs file")
	)
	flag.Parse()

	printTitle()

	// Overview.
	fmt.Printf("\n## Overview\n\n")

	// Handle input variables.
	rawInputs, err := ioutil.ReadFile(*variablesFile)
	if err != nil {
		log.Fatalf("Error reading variables file %q: %s.", *variablesFile, err)
	}

	var hclInput interface{}
	if err := hcl.Unmarshal(rawInputs, &hclInput); err != nil {
		log.Fatalf("Error unmarshalling input: %s", err)
	}

	vars, ok := hclInput.(map[string]interface{})["variable"]
	if !ok && *verbose {
		log.Printf("No variables detected.")
	}

	hclVars := make([]HCLVar, len(vars.([]map[string]interface{})))
	var desc, varType, defaultVal string
	for varindex, varmap := range vars.([]map[string]interface{}) {
		for name, v := range varmap {
			for _, x := range v.([]map[string]interface{}) {
				desc, _ = x["description"].(string)
				varType, _ = x["type"].(string)
				defaultVal, _ = x["default"].(string)
				hclvar := HCLVar{
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
		log.Fatalf("Error templating input: %s", err)
	}
	fmt.Printf("\n## Input\n\n")
	fmt.Println("| Name | Description | Type | Default | Required |")
	fmt.Println("|------|-------------|:----:|:-----:|:-----:|")
	for _, hclvar := range hclVars {
		if err := inputTmpl.Execute(os.Stdout, hclvar); err != nil {
			log.Fatalf("Error executing input on template: %s", err)
		}
	}

	// Handle outputs.
	rawOutputs, err := ioutil.ReadFile(*outputsFile)
	if err != nil {
		log.Fatalf("Error reading outputs file %q: %s.", *outputsFile, err)
	}
	var hclOut interface{}
	if err := hcl.Unmarshal(rawOutputs, &hclOut); err != nil {
		log.Fatalf("Error unmarshalling: %s", err)
	}

	outputs, ok := hclOut.(map[string]interface{})["output"]
	if !ok && *verbose {
		log.Printf("No outputs detected.")
	}

	// NOTE: This smells fishy.
	hclOutputs := make([]HCLVar, 0, len(outputs.([]map[string]interface{})))
	var outputDesc string
	var outputIsSensitive bool
	for _, outmap := range outputs.([]map[string]interface{}) {
		for name, v := range outmap {
			for _, x := range v.([]map[string]interface{}) {
				outputDesc, _ = x["description"].(string)
				outputIsSensitive, _ = x["sensitive"].(bool)
				hclvar := HCLVar{
					Name:        name,
					Description: outputDesc,
					Sensitive:   outputIsSensitive,
				}
				hclOutputs = append(hclOutputs, hclvar)
			}
		}
	}

	// Format and print Outputs.
	outputTmpl, err := template.New("hclvar_output").Parse("| {{.Name}} | {{.Description}} |  {{if .Sensitive}} yes {{else}} no {{end}} |\n")
	if err != nil {
		log.Fatalf("Error templating output: %s.", err)
	}
	fmt.Printf("\n## Output\n\n")
	fmt.Println("| Name | Description | Sensitive |")
	fmt.Println("|------|-------------|:----:|")
	for _, out := range hclOutputs {
		if err := outputTmpl.Execute(os.Stdout, out); err != nil {
			log.Fatalf("Error executing output on template: %s", err)
		}
	}

	// Usage.
	fmt.Printf("\n## Usage\n")
	fmt.Printf("\n```\n\n```\n")

	// Troubleshooting.
	fmt.Printf("\n## Troubleshooting\n\n")
}
