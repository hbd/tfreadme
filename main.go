package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/pkg/errors"
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

func printTitle(w io.Writer) error {
	// Construct the title header.
	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Getwd")
	}
	_, err = fmt.Fprintf(w, "\n# %s Terraform Module\n", strings.ToTitle(filepath.Base(wd)))
	return errors.Wrap(err, "write stream")
}

func hclTable(vars []map[string]interface{}) []HCLVar {
	hclVars := make([]HCLVar, 0, len(vars))

	for _, varmap := range vars {
		for name, v := range varmap {
			for _, x := range v.([]map[string]interface{}) {
				var hclVar HCLVar

				hclVar.Name = name
				hclVar.Description, _ = x["description"].(string)
				hclVar.VarType, _ = x["type"].(string)
				hclVar.DefaultVal, _ = x["default"].(string)
				hclVar.Required = hclVar.DefaultVal == ""
				hclVar.Sensitive, _ = x["sensitive"].(bool)

				hclVars = append(hclVars, hclVar)
			}
		}
	}

	return hclVars
}

func main() {
	var (
		verbose       = flag.Bool("v", false, "verbose mode")
		variablesFile = flag.String("variables", "variables.tf", "path to variables file")
		outputsFile   = flag.String("outputs", "outputs.tf", "path to outputs file")
	)
	flag.Parse()

	w := os.Stdout
	if err := printTitle(w); err != nil {
		log.Fatalf("Error printing title: %s.", err)
	}

	// Overview.
	if _, err := fmt.Fprintf(w, "\n## Overview\n\n"); err != nil {
		log.Fatalf("Error printing overview: %s.", err)
	}

	// Handle input variables.
	rawInputs, err := ioutil.ReadFile(*variablesFile)
	if err != nil {
		log.Fatalf("Error reading variables file %q: %s.", *variablesFile, err)
	}
	var hclInput interface{}
	if err := hcl.Unmarshal(rawInputs, &hclInput); err != nil {
		log.Fatalf("Error unmarshalling input: %s.", err)
	}

	vars, ok := hclInput.(map[string]interface{})["variable"]
	if !ok && *verbose {
		log.Printf("No variables detected.")
	}

	hclVars := hclTable(vars.([]map[string]interface{}))

	// Format and print Inputs.
	inputTmpl, err := template.New("hclvar_input").Parse("| {{.Name}} | {{.Description}} | {{.VarType}} | {{.DefaultVal}} | {{if .Required}} yes {{else}} no {{end}} |\n")
	if err != nil {
		log.Fatalf("Error templating input: %s.", err)
	}
	// TODO: Handle errors.
	_, _ = fmt.Fprintf(w, "\n## Input\n\n")
	_, _ = fmt.Fprintln(w, "| Name | Description | Type | Default | Required |")
	_, _ = fmt.Fprintln(w, "|------|-------------|:----:|:-----:|:-----:|")
	for _, hclvar := range hclVars {
		if err := inputTmpl.Execute(w, hclvar); err != nil {
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
		log.Fatalf("Error unmarshalling: %s.", err)
	}

	outputs, ok := hclOut.(map[string]interface{})["output"]
	if !ok && *verbose {
		log.Printf("No outputs detected.")
	}

	// NOTE: This smells fishy.
	hclOutputs := hclTable(outputs.([]map[string]interface{}))

	// Format and print Outputs.
	outputTmpl, err := template.New("hclvar_output").Parse("| {{.Name}} | {{.Description}} |  {{if .Sensitive}} yes {{else}} no {{end}} |\n")
	if err != nil {
		log.Fatalf("Error templating output: %s.", err)
	}
	// TODO: Handle errors.
	_, _ = fmt.Fprintf(w, "\n## Output\n\n")
	_, _ = fmt.Fprintln(w, "| Name | Description | Sensitive |")
	_, _ = fmt.Fprintln(w, "|------|-------------|:----:|")
	for _, out := range hclOutputs {
		if err := outputTmpl.Execute(w, out); err != nil {
			log.Fatalf("Error executing output on template: %s", err)
		}
	}

	// TODO: Handle errors.
	// Usage.
	_, _ = fmt.Fprintf(w, "\n## Usage\n")
	_, _ = fmt.Fprintf(w, "\n```\n\n```\n")

	// Troubleshooting.
	_, _ = fmt.Fprintf(w, "\n## Troubleshooting\n\n")
}
