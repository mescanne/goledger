package export

import (
	"fmt"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
)

type ExportReport struct {
	JsonPretty bool
	Type       string
}

const export_long = `Export ledger

`

func Add(cmd *cobra.Command, app *app.App, export *ExportReport) {
	ncmd := &cobra.Command{
		Aliases:           []string{},
		Use:               "export [macros|ops...]",
		Short:             "Export transactions",
		Long:              export_long,
		DisableAutoGenTag: true,
	}
	cmd.AddCommand(ncmd)

	// Set defaults
	exportType := utils.NewEnum(&export.Type, []string{"Ledger", "Json", "Beancount", "CSV"}, "exportType")
	ncmd.Flags().Var(exportType, "type", fmt.Sprintf("export type (%s)", exportType.Values()))
	ncmd.Flags().BoolVar(&export.JsonPretty, "jsonpretty", export.JsonPretty, "pretty Json (indented) for Json output")

	// don't need to save it
	macroNames := make([]string, 0, len(app.Macros))
	for k, _ := range app.Macros {
		macroNames = append(macroNames, k)
	}
	ncmd.ValidArgs = macroNames
	ncmd.RunE = func(cmd *cobra.Command, args []string) error {
		return export.run(app, cmd, args)
	}
}

func (export *ExportReport) run(app *app.App, cmd *cobra.Command, args []string) error {

	// Load up saved flags
	b, err := app.LoadBook()
	if err != nil {
		return err
	}

	// Apply ops
	err = app.BookOps(b, args...)
	if err != nil {
		return err
	}

	bp := app.NewBookPrinter(b.GetCCYDecimals())

	// Need type of report now..
	if export.Type == "Json" {
		return bp.PrintJSON(b.Transactions(), export.JsonPretty)
	} else if export.Type == "Beancount" {
		return ShowBeancount(bp, b, app.BaseCCY)
	} else if export.Type == "CSV" {
		return ShowCSV(bp, b)
	} else {
		return ShowLedger(bp, b)
	}
}

