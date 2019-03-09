package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
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
}

const (
	variablesFilename = "variables.tf"
	outputFilename    = "outputs.tf"
)

// exists returns false if the given file does not exist, true otherwise.
func exists(name string) ([]byte, bool) {
	if _, err := os.Stat(variablesFilename); err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
	}
	out, err := ioutil.ReadFile("./" + name)
	if err != nil {
		log.Fatalf("Error reading %s: %v", name, err)
	}
	return out, true
}

func main() {
	// Construct the title header.
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working dir: %s", err)
	}
	fmt.Printf("\n# %s Terraform Module\n", strings.ToTitle(filepath.Base(wd)))

	// Overview.
	fmt.Printf("\n## Overview\n\n")

	// Handle input variables.
	if rawInputs, ok := exists(variablesFilename); ok {
		var hclInput interface{}
		if err := hcl.Unmarshal(rawInputs, &hclInput); err != nil {
			log.Fatalf("Error unmarshalling input: %s", err)
		}

		vars, ok := hclInput.(map[string]interface{})["variable"]
		if !ok {
			log.Fatalf("No variables detected.")
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

		inputTmpl, err := template.New("hclvar_input").Parse("| {{.Name}} | {{.Description}} | {{.VarType}} | {{.DefaultVal}} | {{if .Required}} yes {{else}} no {{end}} |\n")
		if err != nil {
			log.Fatalf("Error templating input: %s", err)
		}
		fmt.Printf("\n## Input\n\n")
		fmt.Println("| Name | Description | Type | Default | Required |")
		fmt.Println("|------|-------------|:----:|:-----:|:-----:|")
		for _, hclvar := range hclVars {
			err = inputTmpl.Execute(os.Stdout, hclvar)
			if err != nil {
				log.Fatalf("Error executing input on template: %s", err)
			}
		}
	}

	// Handle outputs.
	if rawOutputs, ok := exists(outputFilename); ok {
		var hclOut interface{}
		if err := hcl.Unmarshal(rawOutputs, &hclOut); err != nil {
			log.Fatalf("Error unmarshalling: %s", err)
		}

		outputs, ok := hclOut.(map[string]interface{})["output"]
		if !ok {
			log.Fatalf("No variables detected.")
		}

		hclOutputs := make([]HclVar, len(outputs.([]map[string]interface{})))
		var outputDesc, outputType string
		for outindex, outmap := range outputs.([]map[string]interface{}) {
			for name, v := range outmap {
				for _, x := range v.([]map[string]interface{}) {
					outputDesc, _ = x["description"].(string)
					outputType, _ = x["type"].(string)
					hclvar := HclVar{
						Name:        name,
						Description: outputDesc,
						VarType:     outputType,
					}
					hclOutputs[outindex] = hclvar
				}
			}
		}

		outputTmpl, err := template.New("hclvar_output").Parse("| {{.Name}} | {{.Description}} | {{.VarType}} |\n")
		if err != nil {
			log.Fatalf("Error templating output: %s", err)
		}
		fmt.Printf("\n## Output\n\n")
		fmt.Println("| Name | Description | Type |")
		fmt.Println("|------|-------------|:----:|")
		for _, out := range hclOutputs {
			err = outputTmpl.Execute(os.Stdout, out)
			if err != nil {
				log.Fatalf("Error executing output on template: %s", err)
			}
		}
	}

	// Usage.
	fmt.Printf("\n## Usage\n")
	fmt.Printf("\n```\n\n```\n")

	// Troubleshooting.
	fmt.Printf("\n## Troubleshooting\n\n")
}
