package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/mjmar01/harmolytics/internal/helper"
	"github.com/mjmar01/harmolytics/pkg/types"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"text/template"
)

var listFormat string
var knownMethodsFlag bool

const prettyMethod = `
{{range .}}{{.Signature}}:
  Name:   {{.Name}}
  Params: {{.Parameters}}
{{end}}`

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List cached data",
	Args:  cobra.ExactArgs(0),
}

var listMethodsCmd = &cobra.Command{
	Use:     "methods",
	Short:   "List cached methods",
	Args:    cobra.ExactArgs(0),
	PreRunE: openCache,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !helper.StringInSlice(listFormat, []string{"signatures", "json", "pretty"}) {
			return fmt.Errorf("format must be pretty, signatures or json not: %q", listFormat)
		}

		var methods []*types.Method
		switch knownMethodsFlag {
		case true:
			methods = centralCache.GetMethodsByFilter(knownMethods)
		case false:
			methods = centralCache.GetMethodsByFilter(allMethods)
		}
		sort.Slice(methods, func(i, j int) bool {
			return methods[i].Name < methods[j].Name
		})
		switch listFormat {
		case "pretty":
			tmpl, err := template.New("pretty").Parse(prettyMethod)
			cobra.CheckErr(err)
			err = tmpl.Execute(os.Stdout, methods)
			cobra.CheckErr(err)
		case "signatures":
			for _, m := range methods {
				fmt.Println(m.Signature)
			}
		case "json":
			out, err := json.Marshal(methods)
			cobra.CheckErr(err)
			fmt.Println(string(out))
		}
		return nil
	},
}

func allMethods(m *types.Method) bool {
	return true
}

func knownMethods(m *types.Method) bool {
	return m.Name != ""
}

func init() {
	cacheCmd.AddCommand(listCmd)
	listCmd.AddCommand(listMethodsCmd)

	cacheCmd.PersistentFlags().StringVarP(&listFormat, "output-format", "f", "pretty", "json, pretty or signatures")

	listMethodsCmd.PersistentFlags().BoolVarP(&knownMethodsFlag, "known-only", "k", false, "Only show methods where the metadata is known")
}
