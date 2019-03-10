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
	"text/tabwriter"

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

type mdAlign int

const (
	none mdAlign = iota
	left
	right
	center
)

type mdColumn struct {
	Name    string
	Align   mdAlign
	Mapping func(interface{}) interface{}
}

type mdTable struct {
	columns []mdColumn
	rows    [][]interface{}
}

func markdownTable(w io.Writer, table mdTable) error {
	if len(table.columns) < 1 {
		return errors.New("no columns to render")
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug)

	// Print the column names.
	_, _ = fmt.Fprint(tw, "| ")
	for _, c := range table.columns {
		_, _ = fmt.Fprintf(tw, " %s \t", c.Name)
	}
	_, _ = fmt.Fprint(tw, "\n")

	// Print the table header separator.
	_, _ = fmt.Fprint(tw, "|")
	for _, c := range table.columns {
		switch c.Align {
		case none, right:
			_, _ = fmt.Fprint(tw, "-")
		case center, left:
			_, _ = fmt.Fprint(tw, ":")
		}
		_, _ = fmt.Fprintf(tw, "%s", strings.Repeat("-", len(c.Name)))
		switch c.Align {
		case none, left:
			_, _ = fmt.Fprint(tw, "-")
		case center, right:
			_, _ = fmt.Fprint(tw, ":")
		}
		_, _ = fmt.Fprint(tw, "\t")
	}
	_, _ = fmt.Fprint(tw, "\n")

	// Print the rows.
	for _, row := range table.rows {
		_, _ = fmt.Fprint(tw, "|")
		for i, c := range table.columns {
			val := row[i]
			if c.Mapping != nil {
				val = c.Mapping(val)
			}
			_, _ = fmt.Fprintf(tw, " %v \t", val)
		}
		_, _ = fmt.Fprintf(tw, "\n")
	}

	return tw.Flush()
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

	boolFmt := func(in interface{}) interface{} {
		if in.(bool) {
			return "yes"
		}
		return "no"
	}

	_, _ = fmt.Fprintf(w, "\n## Input\n\n")

	variablesTable := mdTable{
		columns: []mdColumn{
			{Name: "Name", Align: none},
			{Name: "Description", Align: left},
			{Name: "Type", Align: center},
			{Name: "Default", Align: center},
			{Name: "Required", Align: center, Mapping: boolFmt},
		},
	}
	for _, hclvar := range hclVars {
		variablesTable.rows = append(variablesTable.rows, []interface{}{
			hclvar.Name,
			hclvar.Description,
			hclvar.VarType,
			hclvar.DefaultVal,
			hclvar.Required,
		})
	}

	if err := markdownTable(w, variablesTable); err != nil {
		log.Fatalf("Error printing variables md table: %s.", err)
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
