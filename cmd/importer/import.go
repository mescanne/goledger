package importer

import (
	"fmt"
	"github.com/antzucaro/matchr"
	"github.com/mescanne/goledger/cmd/app"
	"github.com/mescanne/goledger/cmd/reports"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/spf13/cobra"
	"os"
)

type ImportDef struct {

	// For custom imports
	Name        string
	Description string

	// Decoding of structure (CSV, JSON, etc)
	utils.CLIConfig

	// Code for conversion
	Code string

	// Defaults for code conversion
	Account        string
	CounterAccount string
	CCY            string

	// Rules to apply after import
	Dedup      bool
	Reclassify bool
}

func (imp *ImportDef) run(app *app.App, rcmd *cobra.Command, args []string) error {

	// Use default CCY (base CCY) if non specified for import
	if imp.CCY == "" {
		imp.CCY = app.BaseCCY
	}

	// Create the importer
	importer, err := NewBookImporterByConfig(&imp.CLIConfig)
	if err != nil {
		return fmt.Errorf("invalid import configuration for '%v': %s", &imp.CLIConfig, err)
	}

	// Get the reader
	r := os.Stdin
	if args[0] != "-" {
		r, err = os.Open(args[0])
		if err != nil {
			return fmt.Errorf("error opening %s: %w", args[0], err)
		}
	}

	// Load the records
	imports, err := importer(r)
	if err != nil {
		return fmt.Errorf("error importing: %v", err)
	}

	b, err := imp.processData(imports, imp.Code)
	if err != nil {
		return err
	}

	// Perform deduplication, re-classification if needed
	if imp.Dedup || imp.Reclassify {
		main, err := app.LoadBook()
		if err != nil {
			return fmt.Errorf("error loading book: %s", err)
		}

		if imp.Dedup {
			b.RemoveDuplicatesOf(main)
		}

		if imp.Reclassify {
			b.ReclassifyByAccount(main, imp.Account, func(a, b string) float64 {
				d := matchr.JaroWinkler(a, b, true)
				if d > 0.5 {
					return d
				} else {
					return 0.0
				}
			})
		}
	}

	bp := app.NewBookPrinter(rcmd.OutOrStdout(), b.GetCCYDecimals())

	// Dump report ledger-style
	return reports.ShowLedger(bp, b.Transactions())
}

func (imp *ImportDef) add(name string, app *app.App) *cobra.Command {
	ncmd := &cobra.Command{
		Use:               name,
		Short:             imp.Description,
		Long:              "Import transactions",
		Args:              cobra.ExactArgs(1),
		DisableAutoGenTag: true,
	}
	ncmd.Flags().BoolVarP(&imp.Dedup, "dedup", "d", imp.Dedup, "deduplicate transactions based on payee and date")
	ncmd.Flags().BoolVarP(&imp.Reclassify, "reclassify", "r", imp.Reclassify, "reclassify the counteraccount based on previous transactions")
	if imp.Code == "" {
		ncmd.Flags().StringVar(&imp.Code, "code", imp.Code, "code for import or file:<file> for external code (see help code)")
		cobra.MarkFlagRequired(ncmd.Flags(), "code")
	}
	if imp.ConfigType == "" {
		ncmd.Flags().Var(&imp.CLIConfig, "format", "format of input (see help format)")
		cobra.MarkFlagRequired(ncmd.Flags(), "format")
	}
	ncmd.Flags().StringVarP(&imp.Account, "acct", "a", imp.Account, "account for imported postings")
	if imp.Account == "" {
		cobra.MarkFlagRequired(ncmd.Flags(), "acct")
	}
	ncmd.Flags().StringVarP(&imp.CounterAccount, "cacct", "c", imp.CounterAccount, "counteraccount for new transactions")
	if imp.CounterAccount == "" {
		cobra.MarkFlagRequired(ncmd.Flags(), "cacct")
	}
	ncmd.RunE = func(rcmd *cobra.Command, args []string) error {
		return imp.run(app, rcmd, args)
	}

	return ncmd
}

func Add(root *cobra.Command, app *app.App, config map[string]ImportDef) {

	// Default -- where everything must be configured
	dftl := &ImportDef{
		Description: "Import transactions",
		Dedup:       true,
		Reclassify:  true,
	}
	ncmd := dftl.add("import", app)
	root.AddCommand(ncmd)

	// User customisation -- each customised for each source
	for name, def := range config {
		ncmd.AddCommand(def.add(name, app))
	}

	// Add in the help
	root.AddCommand(&cobra.Command{
		Use:               "format",
		Short:             "Import format help",
		Long:              ImportUsage,
		DisableAutoGenTag: true,
	})
}
